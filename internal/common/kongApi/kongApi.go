package kongApi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const KongAdminURL = "http://localhost:8001"

// Upstream 结构
type Upstream struct {
	Name string `json:"name"`
}

// Target 结构
type Target struct {
	Target string `json:"target"`
	Weight int    `json:"weight"`
}

// Service 结构
type Service struct {
	Name           string `json:"name"`
	Host           string `json:"host"`
	Port           int    `json:"port,omitempty"`
	Protocol       string `json:"protocol,omitempty"`
	Path           string `json:"path,omitempty"`
	Retries        int    `json:"retries,omitempty"`
	ConnectTimeout int    `json:"connect_timeout,omitempty"`
}

// Route 结构
type Route struct {
	Name    string   `json:"name"`
	Paths   []string `json:"paths"`
	Service string   `json:"service"`
}

// 创建 Upstream
func CreateUpstream(name string) error {
	upstream := Upstream{Name: name}
	data, _ := json.Marshal(upstream)

	resp, err := http.Post(KongAdminURL+"/upstreams", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create upstream: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to create upstream: %s", body)
	}
	fmt.Println("Upstream created successfully!")
	return nil
}

// 添加 Target 到 Upstream
func AddTargetToUpstream(upstreamName, target string, weight int) error {
	targetData := Target{
		Target: target,
		Weight: weight,
	}
	data, _ := json.Marshal(targetData)

	resp, err := http.Post(KongAdminURL+"/upstreams/"+upstreamName+"/targets", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to add target to upstream: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to add target to upstream: %s", body)
	}
	fmt.Println("Target added successfully!")
	return nil
}

// 创建 Service
func CreateService(name, hostName, protocol string) error {
	service := Service{
		Name:     name,
		Host:     hostName, // Service 指向 Host/Upstream
		Protocol: protocol, // 根据需求选择 http/https
	}
	data, _ := json.Marshal(service)

	resp, err := http.Post(KongAdminURL+"/services", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to create service: %s", body)
	}
	fmt.Println("Service created successfully!")
	return nil
}

// GetServiceIDByName 根据服务名称获取服务的 ID
func GetServiceIDByName(serviceName string) (string, error) {
	resp, err := http.Get(KongAdminURL + "/services/" + serviceName)
	if err != nil {
		return "", fmt.Errorf("failed to get service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get service: %s", body)
	}

	var serviceData map[string]interface{}
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &serviceData); err != nil {
		return "", fmt.Errorf("failed to unmarshal service data: %v", err)
	}

	// 返回 service.id
	if serviceID, ok := serviceData["id"].(string); ok {
		return serviceID, nil
	}
	return "", fmt.Errorf("service ID not found")
}

// 创建 Route
func CreateRoute(name, serviceName string, paths []string) error {
	// 获取服务的 ID
	serviceID, err := GetServiceIDByName(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service ID: %v", err)
	}

	route := Route{
		Name:    name,
		Paths:   paths,
		Service: serviceID, // 使用服务 ID
	}
	data, _ := json.Marshal(route)

	resp, err := http.Post(KongAdminURL+"/routes", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create route: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to create route: %s", body)
	}
	fmt.Println("Route created successfully!")
	return nil
}
