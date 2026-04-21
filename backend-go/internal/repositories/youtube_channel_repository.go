package repositories

import (
	"backend-go/internal/models"

	"gorm.io/gorm"
)

type YouTubeChannelRepository interface {
	FindByChannelID(channelID string) (*models.YouTubeChannel, error)
	FindByUserID(userID int32) ([]models.YouTubeChannel, error)
	Create(channel *models.YouTubeChannel) error
	Update(channel *models.YouTubeChannel) error
}

type youtubeChannelRepository struct {
	baseRepository
}

func NewYouTubeChannelRepository(db *gorm.DB) YouTubeChannelRepository {
	return &youtubeChannelRepository{baseRepository{db: db}}
}

func (r *youtubeChannelRepository) FindByChannelID(channelID string) (*models.YouTubeChannel, error) {
	var channel models.YouTubeChannel
	if err := r.db.Where("channel_id = ?", channelID).First(&channel).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *youtubeChannelRepository) FindByUserID(userID int32) ([]models.YouTubeChannel, error) {
	var channels []models.YouTubeChannel
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}

func (r *youtubeChannelRepository) Create(channel *models.YouTubeChannel) error {
	return r.db.Create(channel).Error
}

func (r *youtubeChannelRepository) Update(channel *models.YouTubeChannel) error {
	return r.db.Save(channel).Error
}
