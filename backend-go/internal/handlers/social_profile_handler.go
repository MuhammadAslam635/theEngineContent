package handlers

import (
	"net/http"

	"backend-go/internal/dto"
	"backend-go/internal/services"

	"github.com/gin-gonic/gin"
)

type SocialProfileHandler struct {
	service services.SocialProfileService
}

func NewSocialProfileHandler(service services.SocialProfileService) *SocialProfileHandler {
	return &SocialProfileHandler{service: service}
}

// FetchProfile fetches a social media channel profile from SociaVault and saves it.
// @Summary Fetch social media channel profile
// @Tags Social Profiles
// @Accept json
// @Produce json
// @Param body body dto.FetchProfileRequest true "Platform and channel name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /social/fetch-profile [post]
func (h *SocialProfileHandler) FetchProfile(c *gin.Context) {
	var req dto.FetchProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Extract user_id from JWT claims in production
	userID := int32(1)

	channel, err := h.service.FetchAndSaveProfile(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": channel})
}

// GetAllYouTubeChannelsSSE streams all YouTube channels as Server-Sent Events.
// @Summary Get all YouTube channels (SSE)
// @Tags Social Profiles
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /social/youtube/channels [get]
func (h *SocialProfileHandler) GetAllYouTubeChannelsSSE(c *gin.Context) {
	// TODO: Extract user_id from JWT claims in production
	userID := int32(1)

	channels, err := h.service.GetAllYouTubeChannels(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": channels})
}
