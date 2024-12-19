package service

import (
	"errors"
	"fmt"
	"net"
)

type ServiceGo interface {
	InitService() (*Service, error)
	StartGrpcServer() error
	StartHttpServer() error
	Register() error
}

func FindAvailableEndpoint(numOfIp, numOfPort int) ([]string, []int, error) {
	/*
		返回numOfIp个可用Ip地址和numOfPort个可用端口
	*/
	if numOfIp <= 0 || numOfPort <= 0 {
		return nil, nil, errors.New("number of IPs and ports must be greater than 0")
	}

	// Get available IPs from the local network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	ipSet := make(map[string]struct{})
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ipSet[ipNet.IP.String()] = struct{}{}
					if len(ipSet) == numOfPort {
						break
					}
				}
			}
		}
	}
	//fmt.Println(ipSet)
	if len(ipSet) < numOfIp {
		return nil, nil, errors.New("not enough available IPs on the local machine")
	}

	ips := make([]string, 0, numOfIp)
	for ip := range ipSet {
		ips = append(ips, ip)
	}

	// Find available ports
	ports := make([]int, 0, numOfPort)
	for i := 0; i < numOfPort; i++ {
		ln, err := net.Listen("tcp", ":0")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to find an available port: %v", err)
		}
		defer ln.Close()
		ports = append(ports, ln.Addr().(*net.TCPAddr).Port)
	}

	return ips, ports, nil
}
