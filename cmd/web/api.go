package main

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
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

type jwtToken struct {
	ExpiresAt int64 `json:"exp"`
}

func isExpired(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		log.Errorf("no payload found in token")
		return true
	}
	base64Payload := parts[1]
	jsonPayload, err := base64.RawStdEncoding.DecodeString(base64Payload)
	if err != nil {
		log.Error("error decoding payload")
		return true
	}
	var jt jwtToken
	err = json.Unmarshal(jsonPayload, &jt)
	if err != nil {
		log.Error(err)
		return true
	}
	expiry := time.Unix(jt.ExpiresAt, 0)
	now := time.Now()
	thirtySecondsFromNow := now.Add(30 * time.Second)
	if expiry.Before(thirtySecondsFromNow) {
		log.Info("token is expired")
		return true
	}
	seconds := int(expiry.Sub(now).Seconds())
	minutes := seconds / 60
	seconds = seconds - 60 * minutes
	log.Infof("token is valid for the next %d minutes and %d seconds", minutes, seconds)
	return false
}

func (t tokenAuth) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	if tokenString != "" {
		if !isExpired(tokenString) {
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
