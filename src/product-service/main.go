package main

import (
	"context"
	"product-service/handler"
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
		Paths:       []string{"/service/product"},
	})

	s.UpdateOnStart = true
	if err != nil {
		panic(err)
	}

	sm := ss.NewServiceManager(&handler.ProductService{
		Service: s,
	})
	ctx, _ := context.WithTimeout(context.Background(), 200*time.Second)

	sm.StartService(ctx)

}
