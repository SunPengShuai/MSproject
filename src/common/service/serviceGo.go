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
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type ServiceInfo struct {
	Id          string   //服务运行的ID
	Name        string   //服务运行的名称
	Ip          string   //服务运行的IP
	Port        int      //服务运行的端口
	HttpPort    int      //服务运行的http端口
	Protocol    string   //服务协议（默认http）
	Weight      int      //服务权重
	HealthPath  string   //健康检查路径
	ServicePath string   //转发到下游的请求路径
	RoutesName  string   //Kong路由名称
	Paths       []string //kong路由路径
}

type Service struct {
	ServiceInfo ServiceInfo
	stop        chan error
	leaseId     clientv3.LeaseID
	client      *clientv3.Client
	listener    *net.Listener
	grpcClient  *grpc.ClientConn
}
type ServiceManager struct {
	ServiceGo ServiceGo
}

func NewServiceManager(serviceGo ServiceGo) *ServiceManager {
	return &ServiceManager{
		ServiceGo: serviceGo,
	}
}
func (m *ServiceManager) StartService(ctx context.Context) error {
	err := m.ServiceGo.ServiceStart(m)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	if err != nil {
		panic(err)
	}

	go func() {
		select {
		case <-ctx.Done():
			m.ServiceGo.ServiceQuit()

		}
	}()

	sig := <-sigs
	m.ServiceGo.ServiceQuit()
	fmt.Printf("Received signal: %s. Exiting...\n", sig)

	return nil
}

type ServiceGo interface {
	ServiceStart(m *ServiceManager) error
	ServiceQuit() error
	StartGrpcService() (*net.Listener, error)
	StartGrpcGatewayService() (*grpc.ClientConn, error)
	ServiceRegister() error
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

func (s *Service) ServiceStart(m *ServiceManager) error {
	listener, err := m.ServiceGo.StartGrpcService()
	s.listener = listener
	if err != nil {
		log.Panic(err)
	}
	clientConn, err := m.ServiceGo.StartGrpcGatewayService()
	s.grpcClient = clientConn
	if err != nil {
		log.Panic(err)
	}

	if err != nil {
		log.Fatalln("Failed to load swagger spec:", err)
	}

	err = m.ServiceGo.ServiceRegister()
	if err != nil {
		log.Panic(err)
	}
	err = m.ServiceGo.ServiceKong()
	if err != nil {
		log.Panic(err)
	}
	return nil
}
func (s *Service) ServiceQuit() error {
	if err := s.Revoke(context.Background()); err != nil {
		log.Fatalln("error in unregister etcd", err)
	}
	if err := s.UnregisterKong(); err != nil {
		log.Fatalln("error in unregister Kong", err)
	}
	s.grpcClient.Close()
	(*s.listener).Close()
	log.Println("service quit safely")
	return nil
}

func (s *Service) StartGrpcService() (*net.Listener, error) {
	return nil, errors.New("must implement StartGrpcService")
}
func (s *Service) StartGrpcGatewayService() (*grpc.ClientConn, error) {
	return nil, errors.New("must implement StartGrpcGatewayService")
}

// ServiceRegister 注册服务到etcd
func (s *Service) ServiceRegister() error {
	// 注册服务到服务注册中心
	err := RegisterService(s, endpoints)

	if err != nil {
		log.Fatal(err)
		return err
	}
	go s.StartCheckAlive(context.Background())

	log.Println("Service registered successfully")

	return nil
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
	if s.ServiceInfo.HealthPath != "" {

		healthChecks := k.HealthChecks{
			Active: k.ActiveHealthCheck{
				HTTPPath: s.ServiceInfo.HealthPath,
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
		sid, err := k.CreateService(s.ServiceInfo.Name, s.ServiceInfo.Name, s.ServiceInfo.Protocol, s.ServiceInfo.ServicePath)
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

func (s *Service) UnregisterKong() error {
	err := k.UpdateTargetInUpstream(s.ServiceInfo.Name, s.ServiceInfo.Ip+":"+strconv.Itoa(s.ServiceInfo.HttpPort), 0)
	if err != nil {
		log.Fatalf("Error updating target: %v", err)
		return err
	}
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
