package models

type Event struct {
	Type                   string                              `json:"type"`
	DisplayName            string                              `json:"display_name"`
	UserID                 string                              `json:"user_id"`
	Message                string                              `json:"message"`
	IsAnonymous            bool                                `json:"is_anonmyous"`
	Subscription           *EventSubscription                  `json:"subscription,omitempty"`
	ChannelPointRedemption *EventCustomChannelPointsRedemption `json:"channel_point_redemption,omitempty"`
	Bits                   *EventCheer                         `json:"bits,omitempty"`
	Raid                   *EventRaid                          `json:"raid,omitempty"`
}

type EventSubscription struct {
	IsGift bool   `json:"is_gift"`
	Tier   string `json:"tier"`
	Total  int    `json:"total"`
}

type EventCustomChannelPointsRedemption struct {
	RewardID string `json:"reward_id"`
	Title    string `json:"title"`
	Cost     int    `json:"cost"`
	Prompt   string `json:"prompt"`
}

type EventCheer struct {
	Amount int `json:"amount"`
}

type EventRaid struct {
	Viewers int `json:"viewers"`
}
