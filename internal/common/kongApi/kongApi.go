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
	Service struct {
		ID string `json:"id"`
	} `json:"service"`
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

// 创建 Service 并返回服务 ID
func CreateService(name, hostName, protocol, path string) (string, error) {
	service := Service{
		Name:     name,
		Host:     hostName,
		Protocol: protocol,
		Path:     path,
	}
	data, _ := json.Marshal(service)

	resp, err := http.Post(KongAdminURL+"/services", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create service: %s", body)
	}

	var serviceData map[string]interface{}
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &serviceData); err != nil {
		return "", fmt.Errorf("failed to unmarshal service data: %v", err)
	}

	if serviceID, ok := serviceData["id"].(string); ok {
		fmt.Println("Service created successfully with ID:", serviceID)
		return serviceID, nil
	}
	return "", fmt.Errorf("service ID not found")
}

// 创建 Route
func CreateRoute(name, serviceID string, paths []string) error {
	route := Route{
		Name:  name,
		Paths: paths,
	}
	route.Service.ID = serviceID // 绑定服务的 ID

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

// 检查 Target 是否存在
func TargetExists(upstreamName, target string) (bool, error) {
	resp, err := http.Get(KongAdminURL + "/upstreams/" + upstreamName + "/targets")
	if err != nil {
		return false, fmt.Errorf("failed to check target: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return false, fmt.Errorf("failed to check target: %s", body)
	}

	var data map[string]interface{}
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &data); err != nil {
		return false, fmt.Errorf("failed to unmarshal target data: %v", err)
	}

	if targets, ok := data["data"].([]interface{}); ok {
		for _, t := range targets {
			if targetInfo, ok := t.(map[string]interface{}); ok {
				if targetInfo["target"] == target {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// 检查 Upstream 是否存在
func UpstreamExists(name string) (bool, error) {
	return ResourceExists("/upstreams/" + name)
}

// 检查 Route 是否存在
func RouteExists(name string) (bool, error) {
	return ResourceExists("/routes/" + name)
}

// 示例：调用通用方法检查 Service 是否存在
func ServiceExists(name string) (bool, error) {
	return ResourceExists("/services/" + name)
}
func ResourceExists(resourcePath string) (bool, error) {
	resp, err := http.Get(KongAdminURL + resourcePath)
	if err != nil {
		return false, fmt.Errorf("failed to check resource: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return false, fmt.Errorf("error checking resource: %s", body)
	}
	return true, nil
}

func GetServiceID(serviceName string) (string, error) {
	// 构造请求 URL
	url := fmt.Sprintf(KongAdminURL+"/services/%s", serviceName)

	// 发送 HTTP GET 请求
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	res := map[string]interface{}{}
	json.Unmarshal(body, &res)

	return res["id"].(string), nil
}
