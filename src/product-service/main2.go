package main

import (
	"context"
	"product-service/handler"
	"product-service/models"
	ss "service"
	"time"
)

func main() {

	s, err := ss.NewService(&ss.ServiceInfo{
		Name:        "product", //Service和Upstream的名称
		Weight:      100,
		RoutesName:  "product-route",
		Protocol:    "http",
		HealthPath:  "/health",
		ServicePath: "/products",
		Paths:       []string{"/service/products"},
	})

	s.GormMigrate("root:root@tcp(127.0.0.1:3307)/msmall?charset=utf8mb4&parseTime=True&loc=Local", &models.Product{})

	s.UpdateOnStart = true
	if err != nil {
		panic(err)
	}

	sm := ss.NewServiceManager(&handler.ProductService{
		Service: s,
	})

	ctx, _ := context.WithTimeout(context.Background(), 200*time.Minute)

	sm.StartService(ctx)

}
