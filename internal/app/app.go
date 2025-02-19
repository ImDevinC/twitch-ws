package app

import (
	"context"
	"encoding/json"
	"maps"
	"os"
	"os/signal"
	"syscall"

	"github.com/imdevinc/twitch-ws/internal/models"
	"github.com/imdevinc/twitch-ws/internal/sockets"
	"github.com/joeyak/go-twitch-eventsub/v3"
	"github.com/sirupsen/logrus"
)

type app struct {
	clientID        string
	accessToken     string
	userID          string
	subscriptionURL string
	logger          *logrus.Logger
	client          *twitch.Client
	socket          *sockets.Server
}

func Start(logger *logrus.Logger, config Config) error {
	client := twitch.NewClientWithUrl(config.WebsocketURL)
	s := sockets.New(logger, config.Port)
	a := &app{
		clientID:        config.ClientID,
		accessToken:     config.AccessToken,
		userID:          config.UserId,
		logger:          logger,
		client:          client,
		socket:          s,
		subscriptionURL: config.SubscriptionURL,
	}
	client.OnError(func(err error) {
		a.logger.WithError(err).Error("twitch client error")
	})
	client.OnKeepAlive(func(message twitch.KeepAliveMessage) {
		a.logger.WithField("message", message).Debug("keep alive")
	})
	client.OnRevoke(func(message twitch.RevokeMessage) {
		a.logger.WithField("message", message).Error("revoked")
	})
	client.OnWelcome(a.onWelcome)
	client.OnEventChannelChatMessage(a.onChannelChatMessage)
	client.OnEventChannelChannelPointsCustomRewardRedemptionAdd(a.onChannelPointRedemption)
	client.OnEventChannelSubscribe(a.onChannelSubscribe)
	client.OnEventChannelSubscriptionGift(a.onChannelGiftSubscribe)
	client.OnEventChannelSubscriptionMessage(a.onChannelResubscribe)
	client.OnEventChannelFollow(a.onChannelFollow)
	client.OnEventChannelCheer(a.onChannelCheer)
	client.OnEventChannelRaid(a.onChannelRaid)

	go func() {
		if err := client.Connect(); err != nil {
			logger.WithError(err).Error("failed to initialize client")
		}
	}()
	go s.Start()
	defer client.Close()
	defer s.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	return nil
}

func (a *app) sendMessage(event models.Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		a.logger.WithError(err).Error("failed to marshal message")
		return
	}
	a.socket.BroadcastMessage(payload)
}

func (a *app) onWelcome(message twitch.WelcomeMessage) {
	a.logger.WithField("session id", message.Payload.Session.ID).Info("Welcome")
	events := map[twitch.EventSubscription]map[string]string{
		twitch.SubChannelChatMessage: {
			"user_id": a.userID,
		},
		twitch.SubChannelFollow: {
			"moderator_user_id": a.userID,
		},
		twitch.SubChannelSubscribe:           nil,
		twitch.SubChannelSubscriptionGift:    nil,
		twitch.SubChannelSubscriptionMessage: nil, // Resubscribe
		twitch.SubChannelCheer:               nil,
		twitch.SubChannelRaid: {
			"to_broadcaster_user_id": a.userID,
		},
		twitch.SubChannelChannelPointsCustomRewardRedemptionAdd: nil,
		twitch.SubStreamOnline:  nil,
		twitch.SubStreamOffline: nil,
	}
	for event, addParams := range events {
		a.logger.WithField("event", event).Info("subscribing to event")
		params := map[string]string{
			"broadcaster_user_id": a.userID,
		}
		maps.Copy(params, addParams)
		_, err := twitch.SubscribeEventUrlWithContext(context.Background(), twitch.SubscribeRequest{
			SessionID:   message.Payload.Session.ID,
			ClientID:    a.clientID,
			AccessToken: a.accessToken,
			Event:       event,
			Condition:   params,
		}, a.subscriptionURL)
		if err != nil {
			a.logger.WithError(err).WithField("event", event).Error("subscription failed")
		}
	}
}

func (a *app) onChannelFollow(message twitch.EventChannelFollow) {
	a.logger.WithField("message", message).Info("follow")
	event := models.Event{
		DisplayName: message.UserName,
		UserID:      message.UserID,
		Type:        "follow",
	}
	a.sendMessage(event)
}

func (a *app) onChannelGiftSubscribe(message twitch.EventChannelSubscriptionGift) {
	a.logger.WithField("message", message).Info("gift subscription")
	event := models.Event{
		DisplayName: message.UserName,
		UserID:      message.UserID,
		Type:        "gift_sub",
		IsAnonymous: message.IsAnonymous,
		Subscription: &models.EventSubscription{
			Tier:  message.Tier,
			Total: message.Total,
		},
	}
	a.sendMessage(event)
}

func (a *app) onChannelSubscribe(message twitch.EventChannelSubscribe) {
	a.logger.WithField("message", message).Info("subscribe")
	event := models.Event{
		DisplayName: message.UserName,
		UserID:      message.UserID,
		Type:        "subscribe",
		Subscription: &models.EventSubscription{
			Tier:   message.Tier,
			IsGift: message.IsGift,
		},
	}
	a.sendMessage(event)
}

func (a *app) onChannelResubscribe(message twitch.EventChannelSubscriptionMessage) {
	a.logger.WithField("message", message).Info("subscribe with message")
	event := models.Event{
		DisplayName: message.UserName,
		UserID:      message.UserID,
		Type:        "resubscribe",
		Message:     message.Message.Text,
		Subscription: &models.EventSubscription{
			Tier:  message.Tier,
			Total: message.CumulativeMonths,
		},
	}
	a.sendMessage(event)
}

func (a *app) onChannelChatMessage(message twitch.EventChannelChatMessage) {
	a.logger.WithField("message", message).Info("chat message")
	event := models.Event{
		DisplayName: message.ChatterUserName,
		UserID:      message.ChatterUserId,
		Message:     message.Message.Text,
		Type:        "chat",
	}
	a.sendMessage(event)
}

func (a *app) onChannelPointRedemption(message twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
	a.logger.WithField("message", message).Info("channel point redemption")
	event := models.Event{
		DisplayName: message.UserName,
		UserID:      message.UserID,
		Message:     message.UserInput,
		Type:        "channel_points",
		ChannelPointRedemption: &models.EventCustomChannelPointsRedemption{
			RewardID: message.Reward.ID,
			Title:    message.Reward.Title,
			Prompt:   message.Reward.Prompt,
			Cost:     message.Reward.Cost,
		},
	}
	a.sendMessage(event)
}

func (a *app) onChannelCheer(message twitch.EventChannelCheer) {
	a.logger.WithField("message", message).Info("cheer")
	event := models.Event{
		DisplayName: message.UserName,
		UserID:      message.UserID,
		Message:     message.Message,
		Type:        "bits",
		IsAnonymous: message.IsAnonymous,
		Bits: &models.EventCheer{
			Amount: message.Bits,
		},
	}
	a.sendMessage(event)
}

func (a *app) onChannelRaid(message twitch.EventChannelRaid) {
	a.logger.WithField("message", message).Info("raid")
	event := models.Event{
		DisplayName: message.FromBroadcasterUserName,
		UserID:      message.FromBroadcasterUserId,
		Type:        "raid",
		Raid: &models.EventRaid{
			Viewers: message.Viewers,
		},
	}
	a.sendMessage(event)
}
