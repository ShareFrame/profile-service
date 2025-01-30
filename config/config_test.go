package config

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
)

type MockSecretsManager struct {
	GetSecretValueFunc func(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

func (m *MockSecretsManager) GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m.GetSecretValueFunc(ctx, input, opts...)
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		setEnv      bool
		expectError bool
	}{
		{
			"valid environment variable",
			true,
			false,
		},
		{
			"missing environment variable",
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv("ATPROTO_BASE_URL", "https://example.com")
			} else {
				os.Unsetenv("ATPROTO_BASE_URL")
			}

			cfg, _, err := LoadConfig(context.Background())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
			}
		})
	}
}

func TestRetrieveSecret(t *testing.T) {
	tests := []struct {
		name         string
		secretName   string
		mockResponse *secretsmanager.GetSecretValueOutput
		mockError    error
		expectError  bool
	}{
		{
			"successful secret retrieval",
			"my-secret",
			&secretsmanager.GetSecretValueOutput{SecretString: aws.String("super-secret")},
			nil,
			false,
		},
		{
			"missing secret name",
			"",
			nil,
			errors.New("secret name is required"),
			true,
		},
		{
			"secrets manager failure",
			"my-secret",
			nil,
			errors.New("failed to retrieve secret"),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockSecretsManager{
				GetSecretValueFunc: func(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockResponse, nil
				},
			}

			secret, err := RetrieveSecret(context.Background(), tt.secretName, mockSvc)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, *tt.mockResponse.SecretString, secret)
			}
		})
	}
}
