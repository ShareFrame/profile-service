package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/ShareFrame/profile-service/atproto"
	"github.com/ShareFrame/profile-service/config"
	"github.com/ShareFrame/profile-service/models"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/sirupsen/logrus"
)

func retrieveUtilAccountCreds(ctx context.Context, secretsManagerClient config.SecretsManagerAPI) (models.UtilACcountCreds, error) {
	secretName := os.Getenv("PDS_UTIL_ACCOUNT_CREDS")
	input, err := config.RetrieveSecret(ctx, secretName, secretsManagerClient)
	if err != nil {
		logrus.WithError(err).
			WithField("secret_name", secretName).
			Error("Failed to retrieve util account credentials from Secrets Manager")
		return models.UtilACcountCreds{}, fmt.Errorf("error retrieving util account credentials from Secrets Manager (%s): %w", secretName, err)
	}

	var utilAccountCreds models.UtilACcountCreds
	err = json.Unmarshal([]byte(input), &utilAccountCreds)
	if err != nil {
		logrus.WithError(err).
			WithField("secret_content", input).
			Error("Failed to unmarshal util account credentials")
		return models.UtilACcountCreds{}, fmt.Errorf("invalid util account credentials format: %w", err)
	}

	logrus.WithField("username", utilAccountCreds.Username).Info("Successfully retrieved util account credentials")
	return utilAccountCreds, nil
}

func ProfileHandler(ctx context.Context, event map[string]interface{}) (*models.ProfileResponse, error) {
	did, ok := event["did"].(string)
	if !ok || did == "" {
		logrus.Error("Invalid request: missing or empty 'did'")
		return &models.ProfileResponse{}, fmt.Errorf("invalid request: missing or empty 'did'")
	}

	logrus.WithField("did", did).Info("Processing profile retrieval request")

	cfg, awsCfg, err := config.LoadConfig(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to load application configuration")
		return &models.ProfileResponse{}, fmt.Errorf("error loading application configuration: %w", err)
	}

	secretsManagerClient := secretsmanager.NewFromConfig(awsCfg)
	utilAccount, err := retrieveUtilAccountCreds(ctx, secretsManagerClient)
	if err != nil {
		logrus.WithError(err).Error("Failed to retrieve admin credentials from Secrets Manager")
		return &models.ProfileResponse{}, fmt.Errorf("error retrieving admin credentials: %w", err)
	}

	atProtoClient := atproto.NewATProtocolClient(cfg.AtProtoBaseURL, &http.Client{})
	logrus.WithField("base_url", cfg.AtProtoBaseURL).Info("Initializing ATProtocol client")

	session, err := atProtoClient.CreateSession(utilAccount.Username, utilAccount.Password)
	if err != nil {
		logrus.WithError(err).
			WithField("username", utilAccount.Username).
			Error("Failed to create session")
		return &models.ProfileResponse{}, fmt.Errorf("error creating session for user %s: %w", utilAccount.Username, err)
	}

	logrus.WithField("access_token", session.AccessJwt[:10]).Info("Session created successfully (token truncated)")

	profile, err := atProtoClient.GetProfile(ctx, did, session.AccessJwt)
	if err != nil {
		logrus.WithError(err).
			WithField("did", did).
			Error("Failed to get profile")
		return &models.ProfileResponse{}, fmt.Errorf("error retrieving profile for DID %s: %w", did, err)
	}

	logrus.WithFields(logrus.Fields{
		"did":    did,
		"handle": profile.Handle,
	}).Info("Successfully retrieved profile")

	return profile, nil
}
