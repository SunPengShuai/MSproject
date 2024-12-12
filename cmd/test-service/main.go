package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
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
	//启动grpc服务
	lis, err := net.Listen("tcp", ":11234")
	if err != nil {
		log.Fatal(err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()
	pb.RegisterCheckStatusServer(grpcServer, &TestService{})
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
	//注册服务
	//服务发现

}
