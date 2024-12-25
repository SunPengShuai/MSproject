package service

import (
	"context"
	"errors"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	k "kongApi"
	"log"
	"net"
	"strconv"
)

type ServiceInfo struct {
	Id         string   //服务运行的ID
	Name       string   //服务运行的名称
	Ip         string   //服务运行的IP
	Port       int      //服务运行的端口
	HttpPort   int      //服务运行的http端口
	Weight     int      //服务权重
	RoutesName string   //Kong路由名称
	Paths      []string //kong路由路径
}

type Service struct {
	ServiceInfo ServiceInfo
	stop        chan error
	leaseId     clientv3.LeaseID
	client      *clientv3.Client
}
type ServiceManager struct {
	ServiceGo ServiceGo
}

func NewServiceManager(serviceGo ServiceGo) *ServiceManager {
	return &ServiceManager{
		ServiceGo: serviceGo,
	}
}
func (m *ServiceManager) StartService(ctx context.Context) {
	lis, err := m.ServiceGo.StartGrpcService()
	if err != nil {
		fmt.Errorf("ServiceGo.StartGrpcService err:%v", err)
	}
	conn, err := m.ServiceGo.StartGrpcGatewayService()
	if err != nil {
		fmt.Errorf("ServiceGo.StartGrpcGatewayService err:%v", err)
	}
	sev, err := m.ServiceGo.ServiceRegister()
	if err != nil {
		fmt.Errorf("ServiceGo.ServiceRegister err:%v", err)
	}
	err = m.ServiceGo.ServiceKong()
	if err != nil {
		fmt.Errorf("ServiceGo.ServiceKong err:%v", err)
	}
	defer (*lis).Close()
	defer conn.Close()
	defer func() {
		if err := sev.Revoke(context.Background()); err != nil {
			log.Fatalln(err)
		}
	}()
	select {
	case <-ctx.Done():
		return
	}
}

type ServiceGo interface {
	StartGrpcService() (*net.Listener, error)
	StartGrpcGatewayService() (*grpc.ClientConn, error)
	ServiceRegister() (*Service, error)
	ServiceKong() error
}

var endpoints = []string{"127.0.0.1:12379", "127.0.0.1:22379", "127.0.0.1:32379"}

func NewService(serviceInfo *ServiceInfo) (*Service, error) {
	if serviceInfo.Port == 0 || serviceInfo.HttpPort == 0 || serviceInfo.Ip == "" {
		ips, ports, err := FindAvailableEndpoint(1, 2)
		if err != nil {
			log.Fatal("FindAvailableEndpoint err:", err)
		}
		serviceInfo.Port = ports[0]
		serviceInfo.HttpPort = ports[1]
		serviceInfo.Ip = ips[0]
	}
	service := &Service{
		ServiceInfo: *serviceInfo,
	}
	return service, nil
}

// ServiceRegister 注册服务到etcd
func (s *Service) ServiceRegister() (*Service, error) {
	// 注册服务到服务注册中心
	sev, err := RegisterService(s.ServiceInfo, endpoints)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	go sev.StartCheckAlive(context.Background())

	log.Println("Service registered successfully")

	return sev, nil
}

// ServiceKong 注册服务到kong
func (s *Service) ServiceKong() error {

	// 创建 Upstream
	upstreamExists, err := k.UpstreamExists(s.ServiceInfo.Name)
	if err != nil {
		log.Fatalf("Error checking upstream: %v", err)
		return err
	}
	if upstreamExists {
		log.Println("Upstream already exists, updating if needed...")
	} else {
		log.Println("Upstream does not exist, creating...")
		if err := k.CreateUpstream(s.ServiceInfo.Name); err != nil {
			log.Fatalf("Error creating upstream: %v", err)
			return err
		}
	}
	healthChecks := k.HealthChecks{
		Active: k.ActiveHealthCheck{
			HTTPPath: "/health",
			Type:     "http",
			Healthy: k.HealthyStatus{
				HTTPStatuses: []int{200, 201},
				Interval:     5,
			},
			Unhealthy: k.HealthyStatus{
				HTTPStatuses: []int{500, 503},
				Interval:     3,
			},
		},
		Passive: k.PassiveHealthCheck{
			Healthy: k.HealthyStatus{
				HTTPStatuses: []int{200, 201},
			},
			Unhealthy: k.HealthyStatus{
				HTTPStatuses: []int{500, 503},
			},
		},
	}

	// 注册健康检查
	err = k.UpdateHealthChecks(s.ServiceInfo.Name, healthChecks)
	if err != nil {
		log.Fatalf("Error updating health checks: %v", err)
		return err
	}
	// 创建 Target
	targetExists, err := k.TargetExists(s.ServiceInfo.Name, s.ServiceInfo.Ip+":"+strconv.Itoa(s.ServiceInfo.HttpPort))
	if err != nil {
		log.Fatalf("Error checking target: %v", err)
		return err
	}
	if !targetExists {
		log.Println("Target does not exist, adding...")
		if err := k.AddTargetToUpstream(s.ServiceInfo.Name, s.ServiceInfo.Ip+":"+strconv.Itoa(s.ServiceInfo.HttpPort), s.ServiceInfo.Weight); err != nil {
			log.Fatalf("Error adding target: %v", err)
			return err
		}
	}

	// 创建 Service
	serviceExists, err := k.ServiceExists(s.ServiceInfo.Name)
	if err != nil {
		log.Fatalf("Error checking service: %v", err)
		return err
	}

	if serviceExists {
		log.Println("Service already exists, updating if needed...")
		sid, err := k.GetServiceID(s.ServiceInfo.Name)
		if err != nil {
			log.Fatalf("Error getting service ID: %v", err)
			return err
		}
		s.ServiceInfo.Id = sid
	} else {
		log.Println("Service does not exist, creating...")
		sid, err := k.CreateService(s.ServiceInfo.Name, s.ServiceInfo.Name, "http", "/test")
		if err != nil {
			log.Fatalf("Error creating service: %v", err)
			return err
		}
		s.ServiceInfo.Id = sid
	}

	// 创建 Route
	routeExists, err := k.RouteExists(s.ServiceInfo.RoutesName)
	if err != nil {
		log.Fatalf("Error checking route: %v", err)
		return err
	}
	if routeExists {
		log.Println("Route already exists, updating if needed...")
	} else {
		log.Println("Route does not exist, creating...")
		if err := k.CreateRoute(s.ServiceInfo.RoutesName, s.ServiceInfo.Id, s.ServiceInfo.Paths); err != nil {
			log.Fatalf("Error creating route: %v", err)
			return err
		}
	}

	fmt.Println("Service started successfully!")
	return nil
}

func FindAvailableEndpoint(numOfIp, numOfPort int) ([]string, []int, error) {
	/*
		返回numOfIp个可用Ip地址和numOfPort个可用端口
	*/
	if numOfIp <= 0 || numOfPort <= 0 {
		return nil, nil, errors.New("number of IPs and ports must be greater than 0")
	}

	// Get available IPs from the local network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	ipSet := make(map[string]struct{})
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ipSet[ipNet.IP.String()] = struct{}{}
					if len(ipSet) == numOfPort {
						break
					}
				}
			}
		}
	}
	//fmt.Println(ipSet)
	if len(ipSet) < numOfIp {
		return nil, nil, errors.New("not enough available IPs on the local machine")
	}

	ips := make([]string, 0, numOfIp)
	for ip := range ipSet {
		ips = append(ips, ip)
	}

	// Find available ports
	ports := make([]int, 0, numOfPort)
	for i := 0; i < numOfPort; i++ {
		ln, err := net.Listen("tcp", ":0")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to find an available port: %v", err)
		}
		defer ln.Close()
		ports = append(ports, ln.Addr().(*net.TCPAddr).Port)
	}

	return ips, ports, nil
}
