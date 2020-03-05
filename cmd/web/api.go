package main

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"cloud.google.com/go/compute/metadata"
	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	sec "github.com/brymck/risk-service/genproto/brymck/securities/v1"
)

var (
	securitiesApi sec.SecuritiesAPIClient
	tokenString   string
)

type tokenAuth struct {
	serviceName string
}

func (t tokenAuth) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	if tokenString != "" {
		_, _, err := new(jwt.Parser).ParseUnverified(tokenString, &jwt.StandardClaims{})
		expired := false
		switch err.(type) {
		case *jwt.ValidationError:
			vErr := err.(*jwt.ValidationError)
			switch vErr.Errors {
			case jwt.ValidationErrorExpired:
				expired = true
			}
		}

		if !expired {
			return map[string]string{
				"authorization": fmt.Sprintf("Bearer %s", tokenString),
			}, nil
		}
	}

	var newTokenString string
	var err error
	if isOnCloudRun() {
		log.Info("retrieving token from metadata server")
		serviceUrl := getServiceUrl(t.serviceName)
		tokenUrl := fmt.Sprintf("/instance/service-accounts/default/identity?audience=%s", serviceUrl)
		newTokenString, err = metadata.Get(tokenUrl)
		if err != nil {
			return nil, fmt.Errorf("metadata.Get: failed to query id_token: %+v", err)
		}
	} else {
		log.Info("retrieving token from BRYMCK_ID_TOKEN environment variable")
		newTokenString = os.Getenv("BRYMCK_ID_TOKEN")
	}
	if newTokenString == "" {
		return nil, errors.New("token not set")
	}

	tokenString = newTokenString

	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", tokenString),
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
	return fmt.Sprintf("https://%s-4tt23pryoq-an.a.run.app", serviceName)
}

func getGrpcClientConnection(serviceName string) (*grpc.ClientConn, error) {
	pool, _ := x509.SystemCertPool()
	ce := credentials.NewClientTLSFromCert(pool, "")

	conn, err := grpc.Dial(
		getServiceAddress(serviceName),
		grpc.WithTransportCredentials(ce),
		grpc.WithPerRPCCredentials(tokenAuth{serviceName: serviceName}),
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
