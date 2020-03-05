package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	sec "github.com/brymck/risk-service/genproto/brymck/securities/v1"
)

var securitiesApi sec.SecuritiesAPIClient

type tokenAuth struct {
	token string
}

func (t tokenAuth) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", t.token),
	}, nil
}

func (tokenAuth) RequireTransportSecurity() bool {
	return true
}

func getGrpcClientConnection(addr string) (*grpc.ClientConn, error) {
	if strings.Contains(addr, "localhost") {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		return conn, nil
	} else {

		pool, _ := x509.SystemCertPool()
		ce := credentials.NewClientTLSFromCert(pool, "")

		conn, err := grpc.Dial(
			addr,
			grpc.WithTransportCredentials(ce),
			grpc.WithPerRPCCredentials(tokenAuth{token: os.Getenv("BRYMCK_ID_TOKEN")}),
		)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

func getServiceAddress(serviceName string) string {
	return fmt.Sprintf("%s-4tt23pryoq-an.a.run.app:443", serviceName)
}

func init() {
	conn, err := getGrpcClientConnection(getServiceAddress("securities-service"))
	if err != nil {
		log.Fatal(err)
	}
	securitiesApi = sec.NewSecuritiesAPIClient(conn)
}
