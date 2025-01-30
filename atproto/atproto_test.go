package atproto

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestCreateSession(t *testing.T) {
	tests := []struct {
		name           string
		identifier     string
		password       string
		mockResponse   string
		mockStatusCode int
		expectError    bool
	}{
		{
			"successful session creation",
			"user@example.com",
			"password123",
			`{"access_token": "token123"}`,
			http.StatusOK,
			false,
		},
		{
			"failed session creation",
			"user@example.com",
			"password123",
			"",
			http.StatusUnauthorized,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					resp := httptest.NewRecorder()
					resp.WriteHeader(tt.mockStatusCode)
					resp.WriteString(tt.mockResponse)
					return resp.Result(), nil
				},
			}

			client := NewATProtocolClient("http://localhost", mockClient)
			session, err := client.CreateSession(tt.identifier, tt.password)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	tests := []struct {
		name           string
		did            string
		token          string
		mockResponse   string
		mockStatusCode int
		expectError    bool
	}{
		{
			"successful profile retrieval",
			"did:example:1234",
			"valid_token",
			`{"displayName": "John Doe"}`,
			http.StatusOK,
			false,
		},
		{
			"failed profile retrieval",
			"did:example:1234",
			"invalid_token",
			"",
			http.StatusForbidden,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					resp := httptest.NewRecorder()
					resp.WriteHeader(tt.mockStatusCode)
					resp.WriteString(tt.mockResponse)
					return resp.Result(), nil
				},
			}

			client := NewATProtocolClient("http://localhost", mockClient)
			profile, err := client.GetProfile(context.Background(), tt.did, tt.token)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, profile)
			}
		})
	}
}
