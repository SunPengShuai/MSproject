package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"
	"log"
	"math"
	"net"
	"net/http"
	"product-service/models"
	"product-service/pb"
	ss "service"
	"strconv"
	"time"
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
	//gwmux := runtime.NewServeMux()
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true, // 替代 OrigName
				EmitUnpopulated: true, // 替代 EmitDefaults
			},
			UnmarshalOptions: protojson.UnmarshalOptions{},
		}),
	)
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

func (t *ProductService) CreateProduct(ctx context.Context, req *pb.ProductRequest) (*pb.ProductResponse, error) {
	if req == nil {
		return nil, errors.New("invalid request: nil request")
	}

	// 检查必需字段是否存在
	if req.Product.Name == "" || req.Product.Price <= 0 || req.Product.Num < 0 {
		return nil, errors.New("invalid request: missing required fields or invalid values")
	}
	obj := models.Product{
		Name:       req.Product.Name,
		Price:      req.Product.Price,
		Num:        int(req.Product.Num),
		Unit:       req.Product.Unit,
		Pic:        req.Product.Pic,
		Desc:       req.Product.Desc,
		CreateTime: time.Now(),
	}

	// 插入商品记录
	if err := t.Service.GormDB.WithContext(ctx).Create(&obj).Error; err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &pb.ProductResponse{
		Message: "ok",
	}, nil
}

// GetProduct 获取商品信息
func (t *ProductService) GetProduct(ctx context.Context, req *pb.ProductID) (*pb.Product, error) {
	if req == nil || req.Id == 0 {
		return nil, errors.New("invalid request: missing product ID")
	}

	var product models.Product
	if err := t.Service.GormDB.WithContext(ctx).First(&product, req.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		return nil, fmt.Errorf("failed to retrieve product: %w", err)
	}

	return &pb.Product{
		Id:    int64(product.ID),
		Name:  product.Name,
		Price: product.Price,
		Num:   int32(product.Num),
		Unit:  product.Unit,
		Pic:   product.Pic,
		Desc:  product.Desc,
	}, nil
}

// UpdateProduct 更新商品信息
func (t *ProductService) UpdateProduct(ctx context.Context, req *pb.ProductRequest) (*pb.ProductResponse, error) {
	if req == nil || req.Product == nil || req.Product.Id == 0 {
		return nil, errors.New("invalid request: missing product ID or product data")
	}

	// 查找需要更新的商品
	var product models.Product
	// 更新字段
	product.ID = uint(req.Product.Id)
	product.Name = req.Product.Name
	product.Price = req.Product.Price
	product.Num = int(req.Product.Num)
	product.Unit = req.Product.Unit
	product.Pic = req.Product.Pic
	product.Desc = req.Product.Desc

	t.Service.GormDB.WithContext(ctx).Model(&models.Product{}).Where("id", product.ID).Updates(product)

	return &pb.ProductResponse{Message: "ok"}, nil
}

// DeleteProduct 删除商品
func (t *ProductService) DeleteProduct(ctx context.Context, req *pb.ProductID) (*pb.ProductResponse, error) {
	if req == nil || req.Id == 0 {
		return nil, errors.New("invalid request: missing product ID")
	}

	// 删除商品
	if err := t.Service.GormDB.WithContext(ctx).Delete(&models.Product{}, req.Id).Error; err != nil {
		return nil, fmt.Errorf("failed to delete product: %w", err)
	}

	return &pb.ProductResponse{Message: "ok"}, nil
}

// ListProducts 列出所有商品
func (t *ProductService) ListProducts(ctx context.Context, req *pb.ProductRequest) (*pb.ProductListResponse, error) {
	//fmt.Println(req)
	var products []models.Product
	minPrice := req.MinPrice
	maxPrice := int32(math.MaxInt32)
	if req.MaxPrice > 0 {
		maxPrice = req.MaxPrice
	}

	// 查询所有商品
	if err := t.Service.GormDB.WithContext(ctx).Model(&models.Product{}).Where(req.Product).Where("price >= ?", minPrice).Where("price <= ?", maxPrice).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// 转换为 pb.Product
	var pbProducts []*pb.Product
	for _, product := range products {
		pbProducts = append(pbProducts, &pb.Product{
			Id:    int64(product.ID),
			Name:  product.Name,
			Price: product.Price,
			Num:   int32(product.Num),
			Unit:  product.Unit,
			Pic:   product.Pic,
			Desc:  product.Desc,
		})
	}

	return &pb.ProductListResponse{Products: pbProducts}, nil
}
