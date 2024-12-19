package main

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	k "kongApi"
	"log"
	"net"
	"net/http"
	ss "service"
	"strconv"
	pb "test-service/pb"
)

type TestService struct {
	pb.UnimplementedCheckStatusServer
	ss.ServiceInfo
	TargetName string
}

func (t *TestService) GetStatus(ctx context.Context, empty *pb.Empty) (*pb.TestMsg, error) {
	return &pb.TestMsg{
		Msg:    "ok",
		Status: 200,
	}, nil
}

func main() {
	// 定义服务基本信息
	testService := TestService{
		ServiceInfo: ss.ServiceInfo{
			Ip:       "10.211.55.2", //服务运行的IP地址
			Port:     11236,         //grpc服务端口
			Name:     "test",        //Service和Upstream的名称
			HttpPort: 8889,          //http服务端口
		},
		TargetName: "testService1", //target的名称
	}
	// 启动 grpc 服务
	lis, err := net.Listen("tcp", testService.Ip+":"+strconv.Itoa(testService.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()
	pb.RegisterCheckStatusServer(grpcServer, &testService)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("gRPC server is running on port " + strconv.Itoa(testService.Port))

	// 启动 gRPC-Gateway
	conn, err := grpc.Dial(testService.Ip+":"+strconv.Itoa(testService.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}
	defer conn.Close()

	gwmux := runtime.NewServeMux()
	err = pb.RegisterCheckStatusHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    testService.Ip + ":" + strconv.Itoa(testService.HttpPort),
		Handler: gwmux,
	}

	go func() {
		if err := gwServer.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	log.Println("Serving gRPC-Gateway on http://" + gwServer.Addr)

	// 注册服务到服务注册中心
	var endpoints = []string{"127.0.0.1:12379", "127.0.0.1:22379", "127.0.0.1:32379"}

	sev, err := ss.RegisterService(testService.ServiceInfo, endpoints)
	if err != nil {
		log.Fatal(err)
	}
	go sev.StartCheckAlive(context.Background())
	defer func() {
		if err := sev.Revoke(context.Background()); err != nil {
			log.Fatalln(err)
		}
	}()
	log.Println("Service registered successfully")

	// 注册服务到 Kong
	routeName := "test-route"
	target := gwServer.Addr
	weight := 100
	paths := []string{"/service/test"}

	// 创建 Upstream
	upstreamExists, err := k.UpstreamExists(testService.ServiceInfo.Name)
	if err != nil {
		log.Fatalf("Error checking upstream: %v", err)
	}
	if upstreamExists {
		log.Println("Upstream already exists, updating if needed...")
	} else {
		log.Println("Upstream does not exist, creating...")
		if err := k.CreateUpstream(testService.ServiceInfo.Name); err != nil {
			log.Fatalf("Error creating upstream: %v", err)
		}
	}

	// 创建 Target
	targetExists, err := k.TargetExists(testService.ServiceInfo.Name, target)
	if err != nil {
		log.Fatalf("Error checking target: %v", err)
	}
	if !targetExists {
		log.Println("Target does not exist, adding...")
		if err := k.AddTargetToUpstream(testService.ServiceInfo.Name, target, weight); err != nil {
			log.Fatalf("Error adding target: %v", err)
		}
	}

	// 创建 Service
	serviceExists, err := k.ServiceExists(testService.ServiceInfo.Name)
	if err != nil {
		log.Fatalf("Error checking service: %v", err)
	}

	if serviceExists {
		log.Println("Service already exists, updating if needed...")
		sid, err := k.GetServiceID(testService.ServiceInfo.Name)
		if err != nil {
			log.Fatalf("Error getting service ID: %v", err)
		}
		testService.ServiceInfo.Id = sid
	} else {
		log.Println("Service does not exist, creating...")
		sid, err := k.CreateService(testService.ServiceInfo.Name, testService.ServiceInfo.Name, "http", "/test")
		if err != nil {
			log.Fatalf("Error creating service: %v", err)
		}
		testService.ServiceInfo.Id = sid
	}

	// 创建 Route
	routeExists, err := k.RouteExists(routeName)
	if err != nil {
		log.Fatalf("Error checking route: %v", err)
	}
	if routeExists {
		log.Println("Route already exists, updating if needed...")
	} else {
		log.Println("Route does not exist, creating...")
		if err := k.CreateRoute(routeName, testService.ServiceInfo.Id, paths); err != nil {
			log.Fatalf("Error creating route: %v", err)
		}
	}

	fmt.Println("Service started successfully!")
	select {}
}
