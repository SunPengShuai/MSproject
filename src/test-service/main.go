package main

import (
	"context"
	ss "service"
	"test-service/handler"
	"time"
)

func main() {

	s, err := ss.NewService(&ss.ServiceInfo{
		Name:        "test", //Service和Upstream的名称
		Weight:      100,
		RoutesName:  "test-route",
		Protocol:    "http",
		HealthPath:  "/health",
		ServicePath: "/test",
		Paths:       []string{"/service/test", "/service/testB"},
		Port:        50001,
		Ip:          "127.0.0.1",
	})

	s.UpdateOnStart = true
	if err != nil {
		panic(err)
	}

	sm := ss.NewServiceManager(&handler.TestService{
		Service: s,
	})
	ctx, _ := context.WithTimeout(context.Background(), 200*time.Second)

	sm.StartService(ctx)

}
