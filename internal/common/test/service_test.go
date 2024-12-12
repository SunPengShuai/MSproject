package test

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)
import ss "service"

func TestServiceRegister(t *testing.T) {
	//var endpoints = []string{"127.0.0.1:12379", "127.0.0.1:22379", "127.0.0.1:32379"}
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
	//select {}
	sev.Revoke(context.Background())
	t.Log("register service revoke success")
}

func TestServiceDiscovery(t *testing.T) {
	var endpoints = []string{"localhost:12379", "127.0.0.1:22379", "127.0.0.1:32379"}
	ser := ss.NewServiceDiscovery(endpoints)
	defer ser.Close()

	err := ser.WatchService("test")
	if err != nil {
		log.Fatal(err)
	}

	// 监控系统信号，等待 ctrl + c 系统信号通知服务关闭
	c := make(chan os.Signal, 1)
	go func() {
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	}()
	for {
		select {
		case <-time.Tick(10 * time.Second):
			log.Println(ser.GetServices())
		case <-c:
			log.Println("server discovery exit")
			return
		}
	}
}
