package test

import (
	"fmt"
	k "kongApi"
	"testing"
)

func TestKongApi(t *testing.T) {
	// 定义 Upstream、Service 和 Route 的名称
	upstreamName := "example-upstream"
	serviceName := "example-service"
	routeName := "example-route"

	// 定义 Target
	target := "localhost:8080"
	weight := 100

	// 定义 Route 路径
	paths := []string{"/example"}

	// 注册 Upstream
	if err := k.CreateUpstream(upstreamName); err != nil {
		t.Fatalf("Error creating upstream: %v", err)
	}

	// 添加 Target 到 Upstream
	if err := k.AddTargetToUpstream(upstreamName, target, weight); err != nil {
		t.Fatalf("Error adding target: %v", err)
	}

	// 创建 Service
	sid, err := k.CreateService(serviceName, upstreamName, "http", "/")
	if err != nil {
		t.Fatalf("Error creating service: %v", err)
	}

	// 创建 Route
	if err := k.CreateRoute(routeName, sid, paths); err != nil {
		t.Fatalf("Error creating route: %v", err)
	}

	fmt.Println("Service and Route successfully registered!")
}
