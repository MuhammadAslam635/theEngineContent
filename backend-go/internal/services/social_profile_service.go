package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"backend-go/config"
	"backend-go/internal/dto"
	"backend-go/internal/models"
	"backend-go/internal/repositories"
)

type SocialProfileService interface {
	FetchAndSaveProfile(userID int32, req dto.FetchProfileRequest) (*models.YouTubeChannel, error)
	GetAllYouTubeChannels(userID int32) ([]models.YouTubeChannel, error)
}

type socialProfileService struct {
	ytRepo repositories.YouTubeChannelRepository
	cfg    *config.Config
	client *http.Client
}

func NewSocialProfileService(ytRepo repositories.YouTubeChannelRepository, cfg *config.Config) SocialProfileService {
	return &socialProfileService{
		ytRepo: ytRepo,
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *socialProfileService) GetAllYouTubeChannels(userID int32) ([]models.YouTubeChannel, error) {
	return s.ytRepo.FindByUserID(userID)
}

func (s *socialProfileService) FetchAndSaveProfile(userID int32, req dto.FetchProfileRequest) (*models.YouTubeChannel, error) {
	switch req.Platform {
	case "youtube":
		return s.fetchYouTubeProfile(userID, req.ChannelName)
	default:
		return nil, errors.New("platform not supported yet")
	}
}

func (s *socialProfileService) fetchYouTubeProfile(userID int32, channelName string) (*models.YouTubeChannel, error) {
	// Determine if input is a URL or a handle
	var queryParam string
	if strings.HasPrefix(channelName, "http://") || strings.HasPrefix(channelName, "https://") {
		queryParam = "url=" + url.QueryEscape(channelName)
	} else {
		queryParam = "handle=" + url.QueryEscape(channelName)
	}

	// Call SociaVault API
	apiURL := fmt.Sprintf("%syoutube/channel?%s", s.cfg.SociaVaultBaseURL, queryParam)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-API-Key", s.cfg.SociaVaultAPIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call SociaVault: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SociaVault returned %d: %s", resp.StatusCode, string(body))
	}

	var svResp dto.SociaVaultYouTubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&svResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !svResp.Success {
		return nil, errors.New("SociaVault returned unsuccessful response")
	}

	data := svResp.Data

	// Extract avatar URL
	var avatarURL *string
	if src, ok := data.Avatar.Image.Sources["0"]; ok && src.URL != "" {
		avatarURL = &src.URL
	}

	// Convert links map to JSON
	linksJSON, _ := json.Marshal(data.Links)

	// Check if channel already exists — update if so
	existing, err := s.ytRepo.FindByChannelID(data.ChannelID)
	if err == nil && existing != nil {
		// Update existing record
		existing.Name = data.Name
		existing.Handle = strPtr(data.Handle)
		existing.ChannelURL = strPtr(data.Channel)
		existing.AvatarURL = avatarURL
		existing.Description = strPtr(data.Description)
		existing.Country = strPtr(data.Country)
		existing.JoinedDate = strPtr(data.JoinedDateText)
		existing.SubscriberCount = data.SubscriberCount
		existing.SubscriberCountText = strPtr(data.SubscriberCountText)
		existing.VideoCount = data.VideoCount
		existing.ViewCount = data.ViewCount
		existing.ViewCountText = strPtr(data.ViewCountText)
		existing.Tags = strPtr(data.Tags)
		existing.Email = data.Email
		existing.Links = linksJSON
		existing.LastFetchedAt = time.Now()

		if err := s.ytRepo.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// Create new record
	channel := &models.YouTubeChannel{
		UserID:              userID,
		ChannelID:           data.ChannelID,
		ChannelURL:          strPtr(data.Channel),
		Handle:              strPtr(data.Handle),
		Name:                data.Name,
		AvatarURL:           avatarURL,
		Description:         strPtr(data.Description),
		Country:             strPtr(data.Country),
		JoinedDate:          strPtr(data.JoinedDateText),
		SubscriberCount:     data.SubscriberCount,
		SubscriberCountText: strPtr(data.SubscriberCountText),
		VideoCount:          data.VideoCount,
		ViewCount:           data.ViewCount,
		ViewCountText:       strPtr(data.ViewCountText),
		Tags:                strPtr(data.Tags),
		Email:               data.Email,
		Links:               linksJSON,
		LastFetchedAt:       time.Now(),
		IsActive:            true,
	}

	if err := s.ytRepo.Create(channel); err != nil {
		return nil, err
	}
	return channel, nil
}

// strPtr returns a pointer to a string, or nil if empty.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
