package main

import (
	"fmt"
	"strings"

	"github.com/imdevinc/twitch-ws/internal/app"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
)

const wsUrl string = "https://id.twitch.tv/oauth2/authorize?client_id=%s&redirect_uri=http://localhost:7000&response_type=token&scope=%s"

var scopes = []string{
	"bits:read",
	"channel:read:redemptions",
	"channel:read:subscriptions",
	"moderator:read:followers",
	"user:read:chat",
}

func main() {
	logger := logrus.New()
	config, err := app.GetConfigFromEnv()
	if err != nil {
		logger.Fatal(err)
	}

	if config.AccessToken == "" {
		url := fmt.Sprintf(wsUrl, config.ClientID, strings.Join(scopes, "%20"))
		logger.WithField("authorization url", url).Fatal("missing TWITCH_ACCESS_TOKEN. Use the authorization URL if you need to generate one")
	}

	if err := app.Start(logger, app.Config{
		ClientID:        config.ClientID,
		AccessToken:     config.AccessToken,
		UserId:          config.UserId,
		WebsocketURL:    config.WebsocketURL,
		SubscriptionURL: config.SubscriptionURL,
		Port:            config.Port,
	}); err != nil {
		logger.WithError(err).Fatal("app failed")
	}
}
