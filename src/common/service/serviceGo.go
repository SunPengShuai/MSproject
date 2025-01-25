package service

import (
	"context"
	"errors"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	grpc "google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	k "kongApi"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
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
	ServiceInfo    ServiceInfo
	UpdateOnStart  bool
	context        context.Context
	stop           chan error
	leaseId        clientv3.LeaseID
	client         *clientv3.Client
	grpcServer     *grpc.Server
	grpcClientConn *grpc.ClientConn
	listener       net.Listener
	GormDB         *gorm.DB
}

type ServiceManager struct {
	ServiceGo ServiceGo
	Reload    bool //配置热更新的选项
	sigs      chan os.Signal
}

func NewServiceManager(serviceGo ServiceGo) *ServiceManager {
	return &ServiceManager{
		ServiceGo: serviceGo,
	}
}

func (m *ServiceManager) listenForReSet() error {
	for {
		//todo 监听到对应ID的配置文件变动消息
		err := m.ServiceGo.ServiceQuit()
		if err != nil {
			return err
		}
		//todo 拉取最新的配置并更新实体
		//m.ServiceGo.LoadConfig()
		err = m.ServiceGo.ServiceStart(m)
		if err != nil {
			return err
		}
	}
}
func (m *ServiceManager) StopService() error {
	m.ServiceGo.ServiceQuit()
	return nil
}
func (m *ServiceManager) StartService(ctx context.Context) error {
	//m.ServiceGo.SetContext(ctx)
	err := m.ServiceGo.ServiceStart(m)
	m.sigs = make(chan os.Signal, 1)
	signal.Notify(m.sigs, syscall.SIGINT, syscall.SIGTERM)

	if err != nil {
		panic(err)
	}
	if m.Reload {
		go m.listenForReSet()
	}
	select {
	case <-ctx.Done():
		m.StopService()
	case sig := <-m.sigs:
		m.StopService()
		fmt.Printf("Received signal: %s. Exiting...\n", sig)
	}

	return nil
}

type ServiceGo interface {
	LoadConfig(key string) error
	SetContext(ctx context.Context)
	ServiceStart(m *ServiceManager) error
	ServiceQuit() error
	StartGrpcService() (net.Listener, *grpc.Server, error)
	StartGrpcGatewayService() (*grpc.ClientConn, error)
	ServiceRegisterToEtcd() error
	ServiceRegisterToKong() error
}

var endpoints = []string{"127.0.0.1:12379", "127.0.0.1:22379", "127.0.0.1:32379"}

func NewService(serviceInfo *ServiceInfo) (*Service, error) {
	if serviceInfo.Port == 0 || serviceInfo.HttpPort == 0 || serviceInfo.Ip == "" {
		ips, ports, err := FindAvailableEndpoint(1, 2)
		if err != nil {
			log.Fatal("FindAvailableEndpoint err:", err)
		}
		if serviceInfo.Port == 0 {
			serviceInfo.Port = ports[0]
		}
		if serviceInfo.HttpPort == 0 {
			serviceInfo.HttpPort = ports[1]
		}
		if serviceInfo.Ip == "" {
			serviceInfo.Ip = ips[0]
		}
	}
	service := &Service{
		ServiceInfo: *serviceInfo,
		context:     context.Background(),
	}
	return service, nil
}
func (s *Service) LoadConfig(key string) error {
	return nil
}
func (s *Service) SetContext(ctx context.Context) {
	s.context = ctx
}
func (s *Service) ServiceStart(m *ServiceManager) error {
	fmt.Println(s.ServiceInfo)
	listener, grpcserver, err := m.ServiceGo.StartGrpcService()
	s.grpcServer = grpcserver
	s.listener = listener
	if err != nil {
		log.Panic(err)
	}
	clientConn, err := m.ServiceGo.StartGrpcGatewayService()
	s.grpcClientConn = clientConn
	if err != nil {
		log.Panic(err)
	}

	if err != nil {
		log.Fatalln("Failed to load swagger spec:", err)
	}

	err = m.ServiceGo.ServiceRegisterToEtcd()
	if err != nil {
		log.Panic(err)
	}
	err = m.ServiceGo.ServiceRegisterToKong()
	if err != nil {
		log.Panic(err)
	}
	return nil
}
func (s *Service) ServiceQuit() error {
	if err := s.UnregisterKong(); err != nil {
		log.Fatalln("error in unregister Kong", err)
	}
	if err := s.Revoke(context.Background()); err != nil {
		log.Fatalln("error in unregister etcd", err)
	}
	s.grpcServer.GracefulStop() // 关闭 gRPC 服务器
	s.listener.Close()          // 关闭网络监听器
	s.grpcClientConn.Close()    // 关闭 gRPC 客户端连接
	s.client.Close()            // 关闭 etcd 客户端

	log.Println("service quit safely")

	return nil
}
func (s *Service) GormMigrate(dsn string, models ...interface{}) error {
	// 默认 DSN
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/msmall?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// 提取数据库名称和基础 DSN
	dsnWithoutDB := dsn[:strings.LastIndex(dsn, "/")] + "/"
	dbName := dsn[strings.LastIndex(dsn, "/")+1:]
	if idx := strings.Index(dbName, "?"); idx != -1 {
		dbName = dbName[:idx]
	}

	// 连接到 MySQL Server（不包括数据库名）
	serverDB, err := gorm.Open(mysql.Open(dsnWithoutDB), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect to database server: %v\n", err)
		return err
	}

	// 检查数据库是否存在，如果不存在则创建
	createDBQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;", dbName)
	if err := serverDB.Exec(createDBQuery).Error; err != nil {
		fmt.Printf("Failed to create database: %v\n", err)
		return err
	}

	// 重新连接到目标数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetConnMaxLifetime(10 * time.Minute)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return err
	}

	// 自动迁移模型
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			fmt.Printf("Failed to auto migrate: %v\n", err)
		}
	}

	// 保存 DB 实例到 Service
	s.GormDB = db
	return nil
}

func (s *Service) StartGrpcService() (net.Listener, *grpc.Server, error) {
	return nil, nil, errors.New("must implement StartGrpcService")
}
func (s *Service) StartGrpcGatewayService() (*grpc.ClientConn, error) {
	return nil, errors.New("must implement StartGrpcGatewayService")
}

// ServiceRegister 注册服务到etcd
func (s *Service) ServiceRegisterToEtcd() error {
	// 注册服务到服务注册中心
	err := RegisterService(s, endpoints)

	if err != nil {
		log.Fatal(err)
		return err
	}
	go s.StartCheckAlive(s.context)

	log.Println("Service registered successfully")

	return nil
}

// ServiceKong 注册服务到kong
func (s *Service) ServiceRegisterToKong() error {

	// 创建 Upstream
	upstreamExists, err := k.UpstreamExists(s.ServiceInfo.Name)
	if err != nil {
		log.Fatalf("Error checking upstream: %v", err)
		return err
	}
	if upstreamExists && s.UpdateOnStart {
		log.Println("Upstream already exists, updating ...")

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
	} else {
		if s.UpdateOnStart {
			log.Println("Target already exists, updating ...")
			k.UpdateTargetInUpstream(s.ServiceInfo.Name, s.ServiceInfo.Ip+":"+strconv.Itoa(s.ServiceInfo.HttpPort), s.ServiceInfo.Weight)
		}
	}

	// 创建 Service
	serviceExists, err := k.ServiceExists(s.ServiceInfo.Name)
	if err != nil {
		log.Fatalf("Error checking service: %v", err)
		return err
	}

	if serviceExists && s.UpdateOnStart {
		log.Println("Service already exists, updating...")

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
	if routeExists && s.UpdateOnStart {
		log.Println("Route already exists, updating...")

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
		log.Println("Error updating target: %v", err)
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
