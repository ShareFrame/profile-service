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
		logrus.WithField("event", event).Error("Invalid request: missing or empty 'did'")
		return nil, fmt.Errorf("invalid request: missing or empty 'did'")
	}

	logrus.WithField("did", did).Info("Processing profile retrieval request")

	cfg, awsCfg, err := config.LoadConfig(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to load application configuration")
		return nil, fmt.Errorf("internal error: failed to load application configuration")
	}

	secretsManagerClient := secretsmanager.NewFromConfig(awsCfg)
	utilAccount, err := retrieveUtilAccountCreds(ctx, secretsManagerClient)
	if err != nil {
		logrus.WithError(err).Error("Failed to retrieve util account credentials from Secrets Manager")
		return nil, fmt.Errorf("internal error: could not retrieve authentication credentials")
	}

	atProtoClient := atproto.NewATProtocolClient(cfg.AtProtoBaseURL, &http.Client{})
	logrus.WithField("base_url", cfg.AtProtoBaseURL).Info("Initializing ATProtocol client")

	session, err := atProtoClient.CreateSession(utilAccount.Username, utilAccount.Password)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"username": utilAccount.Username,
			"error":    err.Error(),
		}).Error("Failed to create session")
		return nil, fmt.Errorf("authentication failed for user %s", utilAccount.Username)
	}

	logrus.Info("Session created successfully")

	profile, err := atProtoClient.GetProfile(ctx, did, session.AccessJwt)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"did":   did,
			"error": err.Error(),
		}).Error("Failed to retrieve profile")
		return nil, fmt.Errorf("profile not found for DID: %s", did)
	}

	if profile == nil || profile.DID == "" {
		logrus.WithField("did", did).Warn("Profile lookup returned empty result")
		return nil, fmt.Errorf("profile not found for DID: %s", did)
	}

	logrus.WithFields(logrus.Fields{
		"did":         profile.DID,
		"handle":      profile.Handle,
	}).Info("Successfully retrieved profile")

	return profile, nil
}
