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
	input, err := config.RetrieveSecret(ctx, os.Getenv("PDS_UTIL_ACCOUNT_CREDS"), secretsManagerClient)
	if err != nil {
		logrus.WithError(err).
			WithField("secret_name", os.Getenv("PDS_UTIL_ACCOUNT_CREDS")).
			Error("Unable to retrieve util account credentials from Secrets Manager")
		return models.UtilACcountCreds{}, fmt.Errorf("could not retrieve util account credentials from Secrets Manager: %w", err)
	}

	var utilAccountCreds models.UtilACcountCreds
	err = json.Unmarshal([]byte(input), &utilAccountCreds)
	if err != nil {
		logrus.WithError(err).
			WithField("secret_content", input).
			Error("Failed to unmarshal util account credentials")
		return models.UtilACcountCreds{}, fmt.Errorf("invalid util account credentials format in secret: %w", err)
	}

	logrus.Info("Successfully retrieved util account credentials")
	return utilAccountCreds, nil
}

func ProfileHandler(ctx context.Context, event map[string]interface{}) (*models.ProfileResponse, error) {
	did, ok := event["did"].(string)
	if !ok || did == "" {
		logrus.Error("Invalid request: missing 'did'")
		return &models.ProfileResponse{}, fmt.Errorf("invalid request: missing 'did'")
	}

	logrus.WithField("did", did).Info("Processing profile retrieval request")

	cfg, awsCfg, err := config.LoadConfig(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to load application configuration")
		return &models.ProfileResponse{}, fmt.Errorf("could not load application configuration: %w", err)
	}

	secretsManagerClient := secretsmanager.NewFromConfig(awsCfg)
	utilAccount, err := retrieveUtilAccountCreds(ctx, secretsManagerClient)
	if err != nil {
		logrus.WithError(err).Error("Failed to retrieve admin credentials from Secrets Manager")
		return &models.ProfileResponse{}, fmt.Errorf("could not retrieve admin credentials: %w", err)
	}

	atProtoClient := atproto.NewATProtocolClient(cfg.AtProtoBaseURL, &http.Client{})
	session, err := atProtoClient.CreateSession(utilAccount.Username, utilAccount.Password)
	if err != nil {
		logrus.WithError(err).Error("Failed to create session")
		return &models.ProfileResponse{}, fmt.Errorf("could not create session: %w", err)
	}

	profile, err := atProtoClient.GetProfile(ctx, did, session.AccessJwt)
	if err != nil {
		logrus.WithError(err).Error("Failed to get profile")
		return &models.ProfileResponse{}, fmt.Errorf("could not retrieve profile: %w", err)
	}

	logrus.WithField("handle", profile.Handle).Info("Successfully retrieved profile")
	return profile, nil
}
