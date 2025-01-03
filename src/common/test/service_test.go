package test

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"test-service/handler"
	"test-service/pb"
	"testing"
	"time"
)
import ss "service"

func TestServiceRegister(t *testing.T) {

	s, err := ss.NewService(&ss.ServiceInfo{
		Name:        "test", //Service和Upstream的名称
		Weight:      100,
		RoutesName:  "test-route",
		Protocol:    "http",
		HealthPath:  "/health",
		ServicePath: "/test",
		Paths:       []string{"/service/test", "/service/testB"},
	})

	s.UpdateOnStart = true
	if err != nil {
		panic(err)
	}

	sm := ss.NewServiceManager(&handler.TestService{
		Service: s,
	})

	ctx, _ := context.WithTimeout(context.Background(), 200*time.Second)
	go func() {

		select {
		case <-time.After(3 * time.Second):
			sm.StopService()
		}
	}()
	sm.StartService(ctx)

}

func TestServiceDiscovery(t *testing.T) {
	var endpoints = []string{"localhost:12379", "127.0.0.1:22379", "127.0.0.1:32379"}
	ser := ss.NewServiceDiscovery(endpoints)
	defer ser.Close()

	err := ser.WatchService("/services/")
	if err != nil {
		log.Fatal(err)
	}

	// 监控系统信号，等待 ctrl + c 系统信号通知服务关闭
	c := make(chan os.Signal, 1)
	go func() {
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	}()
	for {
		select {
		case <-time.Tick(10 * time.Second):
			log.Println(ser.GetServices())
		case <-c:
			log.Println("server discovery exit")
			return
		}
	}
}

type TestService struct {
	pb.UnimplementedCheckStatusServer
	ss.Service
}

func (t TestService) StartGrpcService() (net.Listener, *grpc.Server, error) {
	// 启动 grpc 服务
	lis, err := net.Listen("tcp", t.ServiceInfo.Ip+":"+strconv.Itoa(t.ServiceInfo.Port))
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCheckStatusServer(grpcServer, &t)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("gRPC server is running on port " + strconv.Itoa(t.ServiceInfo.Port))
	return lis, grpcServer, nil
}
func (t TestService) StartGrpcGatewayService() (*grpc.ClientConn, error) {

	// 启动 gRPC-Gateway
	conn, err := grpc.Dial(t.ServiceInfo.Ip+":"+strconv.Itoa(t.ServiceInfo.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}
	gwmux := runtime.NewServeMux()
	err = pb.RegisterCheckStatusHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    t.ServiceInfo.Ip + ":" + strconv.Itoa(t.ServiceInfo.HttpPort),
		Handler: gwmux,
	}

	go func() {
		if err := gwServer.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	log.Println("Serving gRPC-Gateway on http://" + gwServer.Addr)
	return conn, nil
}

func TestServiceGo(t *testing.T) {
	s, err := ss.NewService(&ss.ServiceInfo{
		Name:       "test", //Service和Upstream的名称
		Weight:     100,
		RoutesName: "test-route",
		Paths:      []string{"/service/test"},
	})

	if err != nil {
		t.Error(err)
	}
	sm := ss.NewServiceManager(&TestService{
		Service: *s,
	})
	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	sm.StartService(ctx)
}
