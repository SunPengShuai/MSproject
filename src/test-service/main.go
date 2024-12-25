package main

import (
	"context"
	ss "service"
)

func main() {
	s, err := ss.NewService(&ss.ServiceInfo{
		Name:       "test", //Service和Upstream的名称
		Weight:     100,
		RoutesName: "test-route",
		Paths:      []string{"/service/test"},
	})

	if err != nil {
		panic(err)
	}
	sm := ss.NewServiceManager(&TestService{
		Service: *s,
	})
	sm.StartService(context.Background())
}