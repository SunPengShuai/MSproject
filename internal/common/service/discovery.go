package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

type Discovery struct {
	endpoints  []string
	service    string
	client     *clientv3.Client
	clientConn resolver.ClientConn
}

func NewDiscovery(endpoints []string, service string) resolver.Builder {
	return &Discovery{
		endpoints: endpoints,
		service:   service,
	}
}

func (d *Discovery) ResolveNow(rn resolver.ResolveNowOptions) {

}

func (d *Discovery) Close() {

}

func (d *Discovery) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var err error
	d.client, err = clientv3.New(clientv3.Config{
		Endpoints: d.endpoints,
	})
	if err != nil {
		return nil, err
	}

	d.clientConn = cc

	go d.watch(d.service)

	return d, nil
}

func (d *Discovery) Scheme() string {
	return "etcd"
}

func (d *Discovery) watch(service string) {
	addrM := make(map[string]resolver.Address)
	state := resolver.State{}

	update := func() {
		addrList := make([]resolver.Address, 0, len(addrM))
		for _, address := range addrM {
			addrList = append(addrList, address)
		}
		state.Addresses = addrList
		err := d.clientConn.UpdateState(state)
		if err != nil {
			fmt.Println("更新地址出错：", err)
		}
	}
	resp, err := d.client.Get(context.Background(), service, clientv3.WithPrefix())
	if err != nil {
		fmt.Println("获取地址出错：", err)
	} else {
		for i, kv := range resp.Kvs {
			info := &ServiceInfo{}
			err = json.Unmarshal(kv.Value, info)
			if err != nil {
				fmt.Println("解析value失败：", err)
			}
			addrM[string(resp.Kvs[i].Key)] = resolver.Address{
				Addr:       info.Ip,
				ServerName: info.Name,
			}
		}
	}

	update()

	dch := d.client.Watch(context.Background(), service, clientv3.WithPrefix(), clientv3.WithPrevKV())
	for response := range dch {
		for _, event := range response.Events {
			switch event.Type {
			case mvccpb.PUT:
				info := &ServiceInfo{}
				err = json.Unmarshal(event.Kv.Value, info)
				if err != nil {
					fmt.Println("监听时解析value报错：", err)
				} else {
					addrM[string(event.Kv.Key)] = resolver.Address{Addr: info.Ip}
				}
				fmt.Println(string(event.Kv.Key))
			case mvccpb.DELETE:
				delete(addrM, string(event.Kv.Key))
				fmt.Println(string(event.Kv.Key))
			}
		}
		update()
	}
}
