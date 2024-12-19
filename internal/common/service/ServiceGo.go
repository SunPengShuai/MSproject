package service

import (
	"google.golang.org/grpc"
	"log"
	"net/http"
)

type BaseService struct {
	GrpcServer *grpc.Server
	HttpServer *http.Server
	EtcdServer *Service
	ServiceGo
}

type ServiceGo interface {
	InitService() (*BaseService, error)
	StartGrpcServer() error
	StartHttpServer() error
	Register() error
}

func (s *BaseService) StartService() error {
	if _, err := s.InitService(); err != nil {
		log.Fatal("service init fail")
		return err
	}
	if err := s.StartGrpcServer(); err != nil {
		log.Fatal("service start grpc fail")
		return err
	}
	if err := s.StartHttpServer(); err != nil {
		log.Fatal("service start http fail")
		return err
	}
	if err := s.Register(); err != nil {
		log.Fatal("service register fail")
		return err
	}
	return nil
}
