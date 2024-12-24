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
	healthChecks := k.HealthChecks{
		Active: k.ActiveHealthCheck{
			HTTPPath: "/health",
			Type:     "http",
			Healthy: k.HealthyStatus{
				HTTPStatuses: []int{200, 201},
				Interval:     5,
			},
			Unhealthy: k.HealthyStatus{
				HTTPStatuses: []int{500, 503},
				Interval:     3,
			},
		},
		Passive: k.PassiveHealthCheck{
			Healthy: k.HealthyStatus{
				HTTPStatuses: []int{200, 201},
				Interval:     10,
			},
			Unhealthy: k.HealthyStatus{
				HTTPStatuses: []int{500, 503},
				Interval:     5,
			},
		},
	}

	// 注册健康检查
	err := k.UpdateHealthChecks(upstreamName, healthChecks)
	if err != nil {
		t.Fatalf("Error updating health checks: %v", err)
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

func TestGetServiceID(t *testing.T) {
	res, err := k.GetServiceID("test")
	if err != nil {
		t.Fatalf("Error getting service ID: %v", err)
	}
	fmt.Println(res)
}
