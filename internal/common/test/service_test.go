package test

import (
	"context"
	"google.golang.org/grpc/resolver"
	"testing"
	"time"
)
import ss "service"

func TestServiceRegister(t *testing.T) {
	var endpoints = []string{"10.4.0.2:2379"}
	serviceInfo := ss.ServiceInfo{
		Name: "test",
		Ip:   "127.0.0.1",
		Port: 8080,
	}
	sev, err := ss.NewService(serviceInfo, endpoints)
	if err != nil {
		t.Error(err)
	}
	go sev.Start(context.Background())
	t.Log("register service success")
	select {
	case <-time.After(time.Second * 2):
		sev.Revoke(context.Background())
	}
	t.Log("register service revoke success")
}

func TestServiceDiscovery(t *testing.T) {
	var endpoints = []string{"10.4.0.2:2379"}
	serviceName := "test"

	// 创建 Discovery 实例
	resolverBuilder := ss.NewDiscovery(endpoints, serviceName)

	// 模拟 gRPC 的 resolver.ClientConn 接口
	mockClientConn := &mockClientConn{
		t: t,
	}

	// 构建解析器
	resolverInstance, err := resolverBuilder.Build(
		resolver.Target{Scheme: "etcd", Endpoint: serviceName},
		mockClientConn,
		resolver.BuildOptions{},
	)
	if err != nil {
		t.Fatalf("Failed to build resolver: %v", err)
	}
	defer resolverInstance.Close()

	// 等待 Discovery 更新地址（根据 watch 实现，可能需要稍作等待）
	time.Sleep(2 * time.Second)

	// 验证获取到的地址是否符合预期
	if len(mockClientConn.state.Addresses) == 0 {
		t.Fatalf("No addresses discovered for service: %s", serviceName)
	}

	for _, addr := range mockClientConn.state.Addresses {
		t.Logf("Discovered address: %s", addr.Addr)
	}
}

// mockClientConn 是一个模拟的 gRPC resolver.ClientConn 实现
type mockClientConn struct {
	t     *testing.T
	state resolver.State
}

func (m *mockClientConn) UpdateState(state resolver.State) error {
	m.state = state // 保存当前的状态以便测试验证
	m.t.Logf("Received updated state: %v", state.Addresses)
	return nil
}

func (m *mockClientConn) ReportError(err error) {
	m.t.Errorf("Reported error: %v", err)
}

func (m *mockClientConn) NewAddress(addresses []resolver.Address) {
	m.t.Logf("Received new addresses: %v", addresses)
}

func (m *mockClientConn) NewServiceConfig(serviceConfig string) {
	m.t.Logf("Received new service config: %v", serviceConfig)
}

func (m *mockClientConn) ParseServiceConfig(serviceConfigJSON string) *resolver.ServiceConfig {
	m.t.Logf("Parse service config called: %v", serviceConfigJSON)
	return &resolver.ServiceConfig{}
}
