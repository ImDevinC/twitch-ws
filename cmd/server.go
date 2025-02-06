package main

import (
	"fmt"
	"net/url"
	"os"
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
	clientID := strings.TrimSpace(os.Getenv("TWITCH_CLIENT_ID"))
	if len(clientID) == 0 {
		logger.Fatal("missing TWITCH_CLIENT_ID")
	}
	userID := strings.TrimSpace(os.Getenv("TWITCH_USER_ID"))
	if len(userID) == 0 {
		logger.Fatal("missing TWITCH_USER_ID")
	}

	// Print this here so that we can get the URL if no access token is available
	authURL := fmt.Sprintf(wsUrl, clientID, url.QueryEscape(strings.Join(scopes, " ")))
	logger.WithField("authURL", authURL).Info("URL")

	accessToken := strings.TrimSpace(os.Getenv("TWITCH_ACCESS_TOKEN"))
	if len(accessToken) == 0 {
		logger.Fatal("missing TWITCH_ACCESS_TOKEN")
	}
	websocketURL := strings.TrimSpace(os.Getenv("TWITCH_WEBSOCKET_URL"))
	if len(websocketURL) == 0 {
		websocketURL = "wss://eventsub.wss.twitch.tv/ws"
	}
	subscriptionURL := strings.TrimSpace(os.Getenv("TWITCH_SUBSCRIPTION_URL"))
	if len(subscriptionURL) == 0 {
		subscriptionURL = "https://api.twitch.tv/helix/eventsub/subscriptions"
	}

	if err := app.Start(logger, app.Config{
		ClientID:        clientID,
		AccessToken:     accessToken,
		UserId:          userID,
		WebsocketURL:    websocketURL,
		SubscriptionURL: subscriptionURL,
		Port:            8080,
	}); err != nil {
		logger.WithError(err).Fatal("app failed")
	}
}
