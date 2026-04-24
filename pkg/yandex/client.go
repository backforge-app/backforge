// Package yandex provides a generic, context-aware HTTP client for interacting
// with the Yandex ID API, specifically focused on the OAuth2 authorization flow
// and retrieving user profile information.
package yandex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	//nolint:gosec
	tokenURL = "https://oauth.yandex.ru/token"
	infoURL  = "https://login.yandex.ru/info?format=json"
)

// UserProfile represents the normalized user data extracted from Yandex.
type UserProfile struct {
	ID              string `json:"id"`
	Email           string `json:"default_email"`
	Name            string `json:"real_name"`
	Login           string `json:"login"`
	DefaultAvatarID string `json:"default_avatar_id"`
	AvatarURL       string `json:"-"` // constructed manually from DefaultAvatarID
}

// Config holds the credentials required for Yandex OAuth.
type Config struct {
	ClientID     string
	ClientSecret string
}

// Client is a generic Yandex API client.
type Client struct {
	cfg        Config
	httpClient *http.Client
}

// New creates a new Yandex API client with sensible production timeouts.
func New(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ExchangeCode executes the full OAuth flow: exchanges the code for a token
// and fetches the profile.
func (c *Client) ExchangeCode(ctx context.Context, code string) (*UserProfile, error) {
	token, err := c.getAccessToken(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	profile, err := c.getUserProfile(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}

	return profile, nil
}

// doJSONRequest is a helper method that executes an HTTP request, safely closes
// the response body, checks the status code, and decodes the JSON response.
func (c *Client) doJSONRequest(req *http.Request, target any) (retErr error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && retErr == nil {
			retErr = fmt.Errorf("close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (c *Client) getAccessToken(ctx context.Context, code string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", c.cfg.ClientID)
	data.Set("client_secret", c.cfg.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := c.doJSONRequest(req, &result); err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", fmt.Errorf("yandex oauth error: %s - %s", result.Error, result.ErrorDesc)
	}

	return result.AccessToken, nil
}

func (c *Client) getUserProfile(ctx context.Context, token string) (*UserProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, infoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "OAuth "+token)

	var profile UserProfile
	if err := c.doJSONRequest(req, &profile); err != nil {
		return nil, err
	}

	if profile.DefaultAvatarID != "" && profile.DefaultAvatarID != "empty-avatar" {
		profile.AvatarURL = fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", profile.DefaultAvatarID)
	}

	return &profile, nil
}
