package handlers

import (
	"net/http"
	"strconv"

	"backend-go/internal/dto"
	"backend-go/internal/services"

	"github.com/gin-gonic/gin"
)

type AgentSettingHandler struct {
	service services.AgentSettingService
}

func NewAgentSettingHandler(service services.AgentSettingService) *AgentSettingHandler {
	return &AgentSettingHandler{service: service}
}

// GetAll returns all agent settings.
// @Summary List all agent settings
// @Tags Agent Settings
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /agents/all [get]
func (h *AgentSettingHandler) GetAll(c *gin.Context) {
	agents, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agents})
}

// GetByID returns a single agent setting by ID.
// @Summary Get agent setting by ID
// @Tags Agent Settings
// @Produce json
// @Param id path int true "Agent Setting ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /agents/{id} [get]
func (h *AgentSettingHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	agent, err := h.service.GetByID(int32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agent})
}

// GetByName returns a single agent setting by agent_name.
// @Summary Get agent setting by name
// @Tags Agent Settings
// @Produce json
// @Param name path string true "Agent Name"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /agents/by-name/{name} [get]
func (h *AgentSettingHandler) GetByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent name is required"})
		return
	}

	agent, err := h.service.GetByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agent})
}

// Create creates a new agent setting.
// @Summary Create agent setting
// @Tags Agent Settings
// @Accept json
// @Produce json
// @Param body body dto.CreateAgentSettingRequest true "Agent Setting"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /agents/create [post]
func (h *AgentSettingHandler) Create(c *gin.Context) {
	var req dto.CreateAgentSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agent, err := h.service.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": agent})
}

// Update updates an existing agent setting.
// @Summary Update agent setting
// @Tags Agent Settings
// @Accept json
// @Produce json
// @Param id path int true "Agent Setting ID"
// @Param body body dto.UpdateAgentSettingRequest true "Fields to update"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /agents/update/{id} [put]
func (h *AgentSettingHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req dto.UpdateAgentSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agent, err := h.service.Update(int32(id), req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agent})
}

// Delete removes an agent setting.
// @Summary Delete agent setting
// @Tags Agent Settings
// @Produce json
// @Param id path int true "Agent Setting ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /agents/delete/{id} [delete]
func (h *AgentSettingHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.Delete(int32(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "agent setting deleted"})
}
