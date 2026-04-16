package httpclient

import (
	"fmt"
	"net/http"
	"time"

	"backend-go/config"
)

type HTTPClient struct {
	client *http.Client
	cfg    *config.Config
}

func NewHTTPClient(cfg *config.Config) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cfg: cfg,
	}
}

func (h *HTTPClient) CheckAIServiceHealth() (int, error) {
	url := fmt.Sprintf("%s/health", h.cfg.AIOrchestrationURL)
	resp, err := h.client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
