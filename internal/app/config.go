package app

import "github.com/kelseyhightower/envconfig"

type Config struct {
	ClientID        string `envconfig:"TWITCH_CLIENT_ID" required:"true"`
	UserId          string `envconfig:"TWITCH_USER_ID" required:"true"`
	AccessToken     string `envconfig:"TWITCH_ACCESS_TOKEN"`
	Port            int    `envconfig:"WS_PORT" default:"8000"`
	WebsocketURL    string `envconfig:"TWITCH_WEBSOCKET_URL" default:"wss://eventsub.wss.twitch.tv/ws"`
	SubscriptionURL string `envconfig:"TWITCH_SUBSCRIPTION_URL" default:"https://api.twitch.tv/helix/eventsub/subscriptions"`
}

func GetConfigFromEnv() (*Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return nil, err
	}
	return &config, nil
}
