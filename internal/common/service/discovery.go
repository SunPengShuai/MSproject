package service

import (
	"context"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"log"
	"sync"
	"time"
)

// ServiceDiscovery 服务发现
type ServiceDiscovery struct {
	cli        *clientv3.Client  // etcd client
	serverList map[string]string // 服务列表
	lock       sync.RWMutex
}

// NewServiceDiscovery 新建服务发现
func NewServiceDiscovery(endpoints []string) *ServiceDiscovery {
	// 初始化etcd client
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Duration(5) * time.Second,
	})
	if err != nil {
		log.Fatalln(err)
	}

	return &ServiceDiscovery{
		cli:        cli,
		serverList: make(map[string]string),
	}
}

// WatchService 初始化服务列表和监视
func (s *ServiceDiscovery) WatchService(prefix string) error {
	// 根据前缀获取现有的key
	resp, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	// 遍历获取得到的k和v
	for _, ev := range resp.Kvs {
		s.SetServiceList(string(ev.Key), string(ev.Value))
	}

	// 监视前缀，修改变更server
	go s.watcher(prefix)
	return nil
}

// watcher 监听Key的前缀
func (s *ServiceDiscovery) watcher(prefix string) {
	rch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Printf("watching prefix:%s now...", prefix)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT: // 修改或者新增
				s.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE: // 删除
				s.DelServiceList(string(ev.Kv.Key))
			}
		}
	}
}

func (s *ServiceDiscovery) SetServiceList(key, val string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.serverList[key] = string(val)
	log.Println("put key:", key, "val:", val)
}

func (s *ServiceDiscovery) DelServiceList(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.serverList, key)
	log.Println("Del key :", key)
}

// GetServices 获取服务地址
func (s *ServiceDiscovery) GetServices() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	addrs := make([]string, 0, len(s.serverList))

	for _, v := range s.serverList {
		addrs = append(addrs, v)
	}
	return addrs
}

// Close 关闭服务
func (s *ServiceDiscovery) Close() error {
	return s.cli.Close()
}
