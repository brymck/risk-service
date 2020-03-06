package auth

import (
	"os"

	"google.golang.org/grpc/credentials"
)

type auth struct {
	credentials.PerRPCCredentials
}

func isOnCloudRun() bool {
	return os.Getenv("K_SERVICE") != ""
}

func NewAuth(tokenUrl string) *auth {
	if isOnCloudRun() {
		return &auth{newMetadataTokenAuth(tokenUrl)}
	} else {
		return &auth{newLocalTokenAuth()}
	}
}
