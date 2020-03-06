package auth

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/compute/metadata"
	log "github.com/sirupsen/logrus"
)

type metadataTokenAuth struct {
	tokenUrl   string
	validUntil time.Time
	token      string
	metadata   map[string]string
}

func newMetadataTokenAuth(tokenUrl string) *metadataTokenAuth {
	beginningOfTime := time.Unix(0, 0)
	return &metadataTokenAuth{tokenUrl: tokenUrl, validUntil: beginningOfTime}
}

func (a *metadataTokenAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	err := a.refresh()
	if err != nil {
		return nil, err
	}
	return a.metadata, nil
}

func (*metadataTokenAuth) RequireTransportSecurity() bool {
	return true
}

func (a *metadataTokenAuth) refresh() error {
	if a.validUntil.Before(time.Now()) {
		log.Debug("retrieving token from metadata server")
		text, err := metadata.Get(a.tokenUrl)
		if err != nil {
			return fmt.Errorf("metadata.Get: failed to query id_token: %+v", err)
		}

		expiresAt, err := getExpiresAt(text)
		if err != nil {
			return err
		}
		a.token = text
		a.validUntil = expiresAt.Add(-timeBeforeExpiryToRenew)
		a.metadata = map[string]string{"authorization": fmt.Sprintf("Bearer %s", a.token)}
	}
	return nil
}
