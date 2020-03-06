package services

import (
	"crypto/x509"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/brymck/risk-service/pkg/auth"
)

type Service interface {
	Connect() (*grpc.ClientConn, error)
	MustConnect() *grpc.ClientConn
}

type service struct {
	name string
}

const (
	serviceAddressTemplate = "%s-4tt23pryoq-an.a.run.app:443"
	serviceUrlTemplate     = "https://%s-4tt23pryoq-an.a.run.app"
	tokenUrlTemplate       = "/instance/service-accounts/default/identity?audience=%s"
)

func NewService(name string) *service {
	return &service{name: name}
}

func (s *service) Connect() (*grpc.ClientConn, error) {
	pool, _ := x509.SystemCertPool()
	ce := credentials.NewClientTLSFromCert(pool, "")

	audience := fmt.Sprintf(serviceUrlTemplate, s.name)
	tokenUrl := fmt.Sprintf(tokenUrlTemplate, audience)
	creds := auth.NewAuth(tokenUrl)

	serviceAddress := fmt.Sprintf(serviceAddressTemplate, s.name)
	conn, err := grpc.Dial(
		serviceAddress,
		grpc.WithTransportCredentials(ce),
		grpc.WithPerRPCCredentials(creds),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *service) MustConnect() *grpc.ClientConn {
	conn, err := s.Connect()
	if err != nil {
		panic(err)
	}
	return conn
}
