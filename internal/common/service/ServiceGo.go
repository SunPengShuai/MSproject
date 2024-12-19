package service

type ServiceGo interface {
	InitService() (*Service, error)
	StartGrpcServer() error
	StartHttpServer() error
	Register() error
}
