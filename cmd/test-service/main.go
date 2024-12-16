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
	pb "test-service/pb"
)

type TestService struct {
	pb.UnimplementedCheckStatusServer
}

func (t *TestService) GetStatus(ctx context.Context, empty *pb.Empty) (*pb.TestMsg, error) {
	return &pb.TestMsg{
		Msg:    "ok",
		Status: 200,
	}, nil
}

func main() {
	//启动grpc/http服务
	lis, err := net.Listen("tcp", ":11234")
	if err != nil {
		log.Fatal(err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()
	pb.RegisterCheckStatusServer(grpcServer, &TestService{})
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("gRPC server is running on port 11234...")

	conn, err := grpc.Dial(
		"localhost:11234",
		grpc.WithInsecure(),
	)
	defer conn.Close()
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = pb.RegisterCheckStatusHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    "0.0.0.0:8888",
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8888")
	go func() {
		if err := gwServer.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	//注册服务
	var endpoints = []string{"127.0.0.1:12379", "127.0.0.1:22379", "127.0.0.1:32379"}
	serviceInfo := ss.ServiceInfo{
		Name: "test",
		Ip:   "127.0.0.1",
		Port: 8080,
	}
	sev, err := ss.NewService(serviceInfo, endpoints)
	if err != nil {
		log.Fatal(err)
	}
	go sev.Start(context.Background())
	log.Println("register service success")
	//select {}
	sev.Revoke(context.Background())
	log.Println("register service revoke success")
	defer func() {
		//服务注销
		err := sev.Revoke(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}()
	//注册服务到kong
	// 定义 Upstream、Service 和 Route 的名称
	upstreamName := "test"
	serviceName := "test-service"
	routeName := "test-route"

	// 定义 Target
	target := gwServer.Addr
	weight := 100

	// 定义 Route 路径
	paths := []string{"/service/test"}
	// 注册 Upstream
	if err := k.CreateUpstream(upstreamName); err != nil {
		fmt.Println("Error creating upstream:", err)
		return
	}
	// 添加 Target 到 Upstream
	if err := k.AddTargetToUpstream(upstreamName, target, weight); err != nil {
		fmt.Println("Error adding target:", err)
		return
	}
	// 创建 Service
	if err := k.CreateService(serviceName, upstreamName, "http"); err != nil {
		fmt.Println("Error creating service:", err)
		return
	}
	// 创建 Route
	if err := k.CreateRoute(66666666, routeName, serviceName, paths); err != nil {
		fmt.Println("Error creating route:", err)
		return
	}

	fmt.Println("Service and Route successfully registered!")
}
