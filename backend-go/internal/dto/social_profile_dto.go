package dto

// FetchProfileRequest is the request to fetch a social media channel profile.
type FetchProfileRequest struct {
	Platform    string `json:"platform" binding:"required,oneof=youtube instagram facebook"`
	ChannelName string `json:"channel_name" binding:"required"`
}

// SociaVaultYouTubeResponse maps the SociaVault YouTube channel API response.
type SociaVaultYouTubeResponse struct {
	Success     bool                      `json:"success"`
	Data        SociaVaultYouTubeData     `json:"data"`
	CreditsUsed int                       `json:"credits_used"`
	Endpoint    string                    `json:"endpoint"`
}

type SociaVaultYouTubeData struct {
	Success             bool                       `json:"success"`
	ChannelID           string                     `json:"channelId"`
	Channel             string                     `json:"channel"`
	Handle              string                     `json:"handle"`
	Name                string                     `json:"name"`
	Avatar              SociaVaultAvatar           `json:"avatar"`
	Description         string                     `json:"description"`
	SubscriberCount     int64                      `json:"subscriberCount"`
	SubscriberCountText string                     `json:"subscriberCountText"`
	VideoCountText      string                     `json:"videoCountText"`
	VideoCount          int                        `json:"videoCount"`
	ViewCountText       string                     `json:"viewCountText"`
	ViewCount           int64                      `json:"viewCount"`
	JoinedDateText      string                     `json:"joinedDateText"`
	Tags                string                     `json:"tags"`
	Email               *string                    `json:"email"`
	Links               map[string]string          `json:"links"`
	Country             string                     `json:"country"`
}

type SociaVaultAvatar struct {
	Image SociaVaultAvatarImage `json:"image"`
}

type SociaVaultAvatarImage struct {
	Sources map[string]SociaVaultImageSource `json:"sources"`
}

type SociaVaultImageSource struct {
	URL string `json:"url"`
}
