package models

type Event struct {
	Type          string `json:"type"`
	DisplayName   string `json:"display_name"`
	UserID        string `json:"user_id"`
	Message       string `json:"message"`
	Bits          int    `json:"bits"`
	ChannelPoints int    `json:"channel_points"`
	IsSub         bool   `json:"is_sub"`
	IsGift        bool   `json:"is_gift"`
	SubTier       string `json:"sub_tier"`
	Viewers       int    `json:"viewers"`
}
