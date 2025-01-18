package main

import (
	"context"
	ss "service"
	"test-service/handler"
	"time"
)

func main() {

	s, err := ss.NewService(&ss.ServiceInfo{
		Name:        "user", //Service和Upstream的名称
		Weight:      100,
		RoutesName:  "user-route",
		Protocol:    "http",
		HealthPath:  "/health",
		ServicePath: "/user",
		Paths:       []string{"/service/userA", "/service/userB"},
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
