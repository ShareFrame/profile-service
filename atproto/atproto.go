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
	logrus.WithField("base_url", baseURL).Info("Initializing ATProtocolClient")
	return &ATProtocolClient{
		BaseURL:    baseURL,
		HTTPClient: client,
	}
}

func (c *ATProtocolClient) doPost(endpoint string, body []byte, headers map[string]string) (*http.Response, error) {
	url := c.BaseURL + endpoint
	logrus.WithFields(logrus.Fields{
		"endpoint": url,
		"headers":  headers,
	}).Debug("Preparing POST request")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		logrus.WithError(err).WithField("url", url).Error("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	logrus.WithField("url", url).Info("Sending POST request")
	return c.HTTPClient.Do(req)
}

func (c *ATProtocolClient) CreateSession(identifier, password string) (*models.SessionResponse, error) {
	logrus.WithField("identifier", identifier).Info("Attempting to create session")
	url := "/xrpc/com.atproto.server.createSession"

	payload := models.SessionRequest{
		Identifier: identifier,
		Password:   password,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal session request payload")
		return nil, fmt.Errorf("failed to marshal session request: %w", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := c.doPost(url, data, headers)
	if err != nil {
		logrus.WithError(err).Error("Failed to execute session creation request")
		return nil, fmt.Errorf("failed to create session request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"url":         url,
		}).Error("Session creation failed")
		return nil, fmt.Errorf("failed to create session, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read session response body")
		return nil, fmt.Errorf("failed to read session response: %w", err)
	}

	var session models.SessionResponse
	if err := json.Unmarshal(body, &session); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal session response")
		return nil, fmt.Errorf("failed to parse session response: %w", err)
	}

	logrus.WithField("identifier", identifier).Info("Session created successfully")
	return &session, nil
}

func (c *ATProtocolClient) GetProfile(ctx context.Context, did string, token string) (*models.ProfileResponse, error) {
	url := fmt.Sprintf("/xrpc/app.bsky.actor.getProfile?actor=%s", did)
	logrus.WithField("did", did).Info("Fetching profile")

	req, err := http.NewRequest("GET", c.BaseURL+url, nil)
	if err != nil {
		logrus.WithError(err).WithField("url", url).Error("Failed to create GET request")
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		logrus.WithError(err).WithField("url", url).Error("Failed to execute profile retrieval request")
		return nil, fmt.Errorf("failed to retrieve profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"url":         url,
		}).Error("Profile retrieval failed")
		return nil, fmt.Errorf("failed to fetch profile, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read profile response body")
		return nil, fmt.Errorf("failed to read profile response: %w", err)
	}

	var profile models.ProfileResponse
	if err := json.Unmarshal(body, &profile); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal profile response")
		return nil, fmt.Errorf("failed to parse profile response: %w", err)
	}

	logrus.WithField("handle", profile.Handle).Info("Successfully retrieved profile")
	return &profile, nil
}
