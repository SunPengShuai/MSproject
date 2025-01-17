package handler

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"product-service/pb"
	ss "service"
	"strconv"
)

type ProductService struct {
	pb.UnimplementedProductServiceServer
	*ss.Service
}

func (t ProductService) StartGrpcService() (net.Listener, *grpc.Server, error) {
	// 启动 grpc 服务
	lis, err := net.Listen("tcp", t.ServiceInfo.Ip+":"+strconv.Itoa(t.ServiceInfo.Port))
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, &t)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("gRPC server is running on " + t.ServiceInfo.Ip + ":" + strconv.Itoa(t.ServiceInfo.Port))
	return lis, grpcServer, nil
}
func (t ProductService) StartGrpcGatewayService() (*grpc.ClientConn, error) {

	// 启动 gRPC-Gateway
	conn, err := grpc.Dial(t.ServiceInfo.Ip+":"+strconv.Itoa(t.ServiceInfo.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}
	gwmux := runtime.NewServeMux()
	err = pb.RegisterProductServiceHandler(context.Background(), gwmux, conn)
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

func (t *ProductService) Health(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (t *ProductService) CreateProduct(context.Context, *pb.ProductRequest) (*pb.ProductResponse, error) {

}
func (t *ProductService) GetProduct(context.Context, *pb.ProductID) (*pb.Product, error) {

}
func (t *ProductService) UpdateProduct(context.Context, *pb.ProductRequest) (*pb.ProductResponse, error) {

}
func (t *ProductService) DeleteProduct(context.Context, *pb.ProductID) (*pb.ProductResponse, error) {

}
func (t *ProductService) ListProducts(context.Context, *pb.Empty) (*pb.ProductListResponse, error) {

}
