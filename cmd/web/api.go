package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	sec "github.com/brymck/risk-service/genproto/brymck/securities/v1"
)

var securitiesApi sec.SecuritiesAPIClient

func getServiceAddress(serviceName string) string {
	return fmt.Sprintf("%s-4tt23pryoq-an.a.run.app:443", serviceName)
}

func init() {
	creds := credentials.NewClientTLSFromCert(nil, "")
	conn, err := grpc.Dial(getServiceAddress("securities-service"), grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal(err)
	}
	securitiesApi = sec.NewSecuritiesAPIClient(conn)
}
