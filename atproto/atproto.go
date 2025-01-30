package atproto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ShareFrame/profile-service/models"
	"github.com/sirupsen/logrus"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ATProtocolClient struct {
	BaseURL    string
	HTTPClient HTTPClient
}

func NewATProtocolClient(baseURL string, client HTTPClient) *ATProtocolClient {
	return &ATProtocolClient{
		BaseURL:    baseURL,
		HTTPClient: client,
	}
}

func (c *ATProtocolClient) doPost(endpoint string, body []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, bytes.NewBuffer(body))
	if err != nil {
		logrus.WithError(err).Error("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	logrus.WithFields(logrus.Fields{
		"endpoint": endpoint,
		"headers":  headers,
	}).Debug("Sending POST request")

	return c.HTTPClient.Do(req)
}

func (c *ATProtocolClient) CreateSession(identifier, password string) (*models.SessionResponse, error) {
	url := "/xrpc/com.atproto.server.createSession"

	payload := models.SessionRequest{
		Identifier: identifier,
		Password:   password,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := c.doPost(url, data, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create session, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var session models.SessionResponse
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (c *ATProtocolClient) GetProfile(ctx context.Context, did string, token string) (*models.ProfileResponse, error) {
	url := fmt.Sprintf("/xrpc/app.bsky.actor.getProfile?actor=%s", did)

	req, err := http.NewRequest("GET", c.BaseURL+url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	client := c.HTTPClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch profile, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var profile models.ProfileResponse
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}
