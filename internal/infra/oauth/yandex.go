// Package oauth provides adapters for third-party OAuth providers,
// mapping their specific responses to our internal application domain.
package oauth

import (
	"context"
	"fmt"

	"github.com/backforge-app/backforge/internal/service/auth"
	"github.com/backforge-app/backforge/pkg/yandex"
)

// YandexAdapter implements the auth.OAuthClient interface.
// It translates the Yandex-specific profile into the application's domain profile.
type YandexAdapter struct {
	client *yandex.Client
}

// NewYandexAdapter creates a new adapter wrapping the generic Yandex client.
func NewYandexAdapter(client *yandex.Client) *YandexAdapter {
	return &YandexAdapter{
		client: client,
	}
}

// ExchangeCode calls the Yandex API and maps the result to auth.OAuthProfile.
func (a *YandexAdapter) ExchangeCode(ctx context.Context, code string) (*auth.OAuthProfile, error) {
	yandexProfile, err := a.client.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("yandex exchange failed: %w", err)
	}

	return &auth.OAuthProfile{
		ProviderID: yandexProfile.ID,
		Email:      yandexProfile.Email,
		Name:       yandexProfile.Name,
		AvatarURL:  yandexProfile.AvatarURL,
	}, nil
}
