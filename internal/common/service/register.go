package service

import (
	"context"
	"encoding/json"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type ServiceInfo struct {
	Id       string
	Name     string
	Ip       string
	Port     int
	HttpPort int
}

type Service struct {
	ServiceInfo ServiceInfo
	stop        chan error
	leaseId     clientv3.LeaseID
	client      *clientv3.Client
}

func RegisterService(serviceInfo ServiceInfo, endpoints []string) (service *Service, err error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 10,
	})
	if err != nil {
		return nil, err
	}

	service = &Service{
		ServiceInfo: serviceInfo,
		client:      client,
	}
	return
}

func (s *Service) StartCheckAlive(ctx context.Context) (err error) {

	alive, err := s.KeepAlive(ctx)
	if err != nil {
		return
	}
	for {
		select {
		case err = <-s.stop: // 服务端关闭返回错误
			return err
		case <-s.client.Ctx().Done(): // etcd关闭
			return errors.New("server closed")
		case _, ok := <-alive:
			if !ok { // 保活通道关闭
				return s.Revoke(ctx)
			}
		}
	}
}

func (s *Service) KeepAlive(ctx context.Context) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	info := s.ServiceInfo
	key := s.getKey()
	val, _ := json.Marshal(info)
	// 创建租约
	leaseResp, err := s.client.Grant(ctx, 5)
	if err != nil {
		return nil, err
	}
	// 写入etcd
	_, err = s.client.Put(ctx, key, string(val), clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return nil, err
	}

	s.leaseId = leaseResp.ID
	return s.client.KeepAlive(ctx, leaseResp.ID)
}

// 取消租约
func (s *Service) Revoke(ctx context.Context) error {
	_, err := s.client.Revoke(ctx, s.leaseId)
	return err
}

func (s *Service) getKey() string {
	return s.ServiceInfo.Name
}
