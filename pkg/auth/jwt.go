package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const timeBeforeExpiryToRenew = 30 * time.Second

type payload struct {
	ExpiresAt int64 `json:"exp"`
}

func decodeJwtToken(token string) (*payload, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, errors.New("no payload found in token")
	}
	base64Payload := parts[1]
	jsonPayload, err := base64.RawStdEncoding.DecodeString(base64Payload)
	if err != nil {
		return nil, err
	}
	var p payload
	err = json.Unmarshal(jsonPayload, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func getExpiresAt(token string) (time.Time, error) {
	jwt, err := decodeJwtToken(token)
	if err != nil {
		return time.Time{}, err
	}
	expiry := time.Unix(jwt.ExpiresAt, 0)
	return expiry, nil
}
