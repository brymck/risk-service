package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

type localTokenAuth struct {
	token      string
	validUntil time.Time
	metadata   map[string]string
}

func newLocalTokenAuth() *localTokenAuth {
	token := os.Getenv("BRYMCK_ID_TOKEN")
	expiry, err := getExpiresAt(token)
	if err != nil {
		panic(err)
	}
	metadata := map[string]string{"authorization": fmt.Sprintf("Bearer %s", token)}
	return &localTokenAuth{token: token, validUntil: expiry, metadata: metadata}
}

func (a *localTokenAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	err := a.refresh()
	if err != nil {
		return nil, err
	}
	return a.metadata, nil
}

func (*localTokenAuth) RequireTransportSecurity() bool {
	return true
}

func (a *localTokenAuth) refresh() error {
	if a.validUntil.Before(time.Now()) {
		return errors.New("expired token")
	}
	return nil
}
