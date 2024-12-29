package handler

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	ss "service"
	"strconv"
	"test-service/pb"
)

type TestService struct {
	pb.UnimplementedCheckStatusServer
	ss.Service
}

func (t TestService) StartGrpcService() (*net.Listener, error) {
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
	return &lis, nil
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
func (t TestService) ServiceKong() error {
	fmt.Println("my logic")
	err := t.Service.ServiceKong()
	fmt.Println("clean something")
	if err != nil {
		return err
	}
	return nil
}
func (t *TestService) GetStatus(ctx context.Context, empty *pb.Empty) (*pb.TestMsg, error) {
	return &pb.TestMsg{
		Msg:    "ok",
		Status: 200,
	}, nil
}

func (t *TestService) GetStatusA(ctx context.Context, empty *pb.Empty) (*pb.TestMsg, error) {
	return &pb.TestMsg{
		Msg:    "service A is ok",
		Status: 200,
	}, nil
}

func (t *TestService) Health(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}
