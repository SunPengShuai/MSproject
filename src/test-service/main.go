package main

import (
	"context"
	ss "service"
	"test-service/handler"
)

func main() {

	s, err := ss.NewService(&ss.ServiceInfo{
		Name:        "test", //Service和Upstream的名称
		Weight:      100,
		RoutesName:  "test-route",
		Protocol:    "http",
		HealthPath:  "/health",
		ServicePath: "/test",
		Paths:       []string{"/service/test"},
	})

	if err != nil {
		panic(err)
	}
	sm := ss.NewServiceManager(&handler.TestService{
		Service: *s,
	})

	sm.StartService(context.Background())

}
