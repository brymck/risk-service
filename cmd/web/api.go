package main

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"cloud.google.com/go/compute/metadata"
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
	log.Infof("providing token %s", t.token)
	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", t.token),
	}, nil
}

func (tokenAuth) RequireTransportSecurity() bool {
	return true
}

func isOnCloudRun() bool {
	if os.Getenv("K_SERVICE") == "" {
		log.Info("not running on Cloud Run because environment variable K_SERVICE is not set")
		return false
	} else {
		log.Info("running on Cloud Run because environment variable K_SERVICE is set")
		return true
	}
}

func getServiceAddress(serviceName string) string {
	return fmt.Sprintf("%s-4tt23pryoq-an.a.run.app:443", serviceName)
}

func getServiceUrl(serviceName string) string {
	return fmt.Sprintf("%s-4tt23pryoq-an.a.run.app", serviceName)
}

func getGrpcClientConnection(serviceName string) (*grpc.ClientConn, error) {
	pool, _ := x509.SystemCertPool()
	ce := credentials.NewClientTLSFromCert(pool, "")

	var token string
	var err error
	if isOnCloudRun() {
		log.Info("retrieving token from metadata server")
		serviceUrl := getServiceUrl(serviceName)
		tokenUrl := fmt.Sprintf("/instance/service-accounts/default/identity?audience=%s", serviceUrl)
		token, err = metadata.Get(tokenUrl)
		if err != nil {
			return nil, fmt.Errorf("metadata.Get: failed to query id_token: %+v", err)
		}
	} else {
		log.Info("retrieving token from BRYMCK_ID_TOKEN environment variable")
		token = os.Getenv("BRYMCK_ID_TOKEN")
	}
	if token == "" {
		return nil, errors.New("token not set")
	}

	conn, err := grpc.Dial(
		getServiceAddress(serviceName),
		grpc.WithTransportCredentials(ce),
		grpc.WithPerRPCCredentials(tokenAuth{token: token}),
	)
	if err != nil {
			  return nil, err
			  }
	return conn, nil
}

func init() {
	conn, err := getGrpcClientConnection("securities-service")
	if err != nil {
		log.Fatal(err)
	}
	securitiesApi = sec.NewSecuritiesAPIClient(conn)
}
