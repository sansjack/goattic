package auth

import (
	"context"
	"errors"
)

var ErrInvalidAPIKey = errors.New("invalid API key")
var ErrMissingAPIKey = errors.New("missing API key")

type APIKeyValidator interface {
	ValidateAPIKey(ctx context.Context, apiKey string) error
}

func ExtractAPIKey(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrMissingAPIKey
	}

	return authHeader, nil
}
