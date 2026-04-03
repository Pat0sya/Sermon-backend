package handlers

import (
	"net/http"
	"time"

	"sermon-backend/internal/models"
	"sermon-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ServerHandler struct {
	DB *gorm.DB
}

type CreateServerRequest struct {
	Name        string `json:"name" binding:"required"`
	Host        string `json:"host" binding:"required"`
	OS          string `json:"os" binding:"required"`
	Description string `json:"description"`
}

type UpdateServerRequest struct {
	Name        string `json:"name" binding:"required"`
	Host        string `json:"host" binding:"required"`
	OS          string `json:"os" binding:"required"`
	Description string `json:"description"`
}

func NewServerHandler(db *gorm.DB) *ServerHandler {
	return &ServerHandler{DB: db}
}

func (h *ServerHandler) GetServers(c *gin.Context) {
	var servers []models.Server
	if err := h.DB.Order("id asc").Find(&servers).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to fetch servers")
		return
	}

	response := make([]gin.H, 0, len(servers))
	for _, s := range servers {
		response = append(response, gin.H{
			"id":           s.ID,
			"name":         s.Name,
			"host":         s.Host,
			"os":           s.OS,
			"description":  s.Description,
			"is_active":    s.IsActive,
			"status":       s.Status,
			"last_seen_at": s.LastSeenAt,
			"created_at":   s.CreatedAt,
		})
	}

	utils.Success(c, http.StatusOK, response)
}

func (h *ServerHandler) GetServerByID(c *gin.Context) {
	var server models.Server
	if err := h.DB.First(&server, c.Param("id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "server not found")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"id":           server.ID,
		"name":         server.Name,
		"host":         server.Host,
		"os":           server.OS,
		"description":  server.Description,
		"is_active":    server.IsActive,
		"status":       server.Status,
		"last_seen_at": server.LastSeenAt,
		"created_at":   server.CreatedAt,
	})
}

func (h *ServerHandler) CreateServer(c *gin.Context) {
	var req CreateServerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Name) < 2 {
		utils.Error(c, http.StatusBadRequest, "name must be at least 2 characters")
		return
	}

	if len(req.Host) < 3 {
		utils.Error(c, http.StatusBadRequest, "host must be at least 3 characters")
		return
	}

	var existing models.Server
	if err := h.DB.Where("host = ?", req.Host).First(&existing).Error; err == nil {
		utils.Error(c, http.StatusConflict, "server with this host already exists")
		return
	}

	agentToken, err := utils.GenerateAgentToken()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to generate agent token")
		return
	}

	agentTokenHash, err := utils.HashAgentToken(agentToken)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to hash agent token")
		return
	}

	server := models.Server{
		Name:           req.Name,
		Host:           req.Host,
		OS:             req.OS,
		Description:    req.Description,
		IsActive:       true,
		Status:         "normal",
		AgentTokenHash: agentTokenHash,
	}

	if err := h.DB.Create(&server).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to create server")
		return
	}
	defaultThresholds := []models.Threshold{
		{ServerID: server.ID, MetricType: "cpu", WarningValue: 80, CriticalValue: 90},
		{ServerID: server.ID, MetricType: "ram", WarningValue: 80, CriticalValue: 90},
		{ServerID: server.ID, MetricType: "disk", WarningValue: 85, CriticalValue: 95},
	}

	if err := h.DB.Create(&defaultThresholds).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to create default thresholds")
		return
	}

	utils.Success(c, http.StatusCreated, gin.H{
		"id":           server.ID,
		"name":         server.Name,
		"host":         server.Host,
		"os":           server.OS,
		"description":  server.Description,
		"is_active":    server.IsActive,
		"status":       server.Status,
		"last_seen_at": server.LastSeenAt,
		"created_at":   server.CreatedAt,
		"agent_token":  agentToken,
	})
}
func (h *ServerHandler) RegenerateAgentToken(c *gin.Context) {
	var server models.Server
	if err := h.DB.First(&server, c.Param("id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "server not found")
		return
	}

	agentToken, err := utils.GenerateAgentToken()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to generate agent token")
		return
	}

	agentTokenHash, err := utils.HashAgentToken(agentToken)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to hash agent token")
		return
	}

	server.AgentTokenHash = agentTokenHash

	if err := h.DB.Save(&server).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to update agent token")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"message":     "agent token regenerated successfully",
		"server_id":   server.ID,
		"agent_token": agentToken,
	})
}

func (h *ServerHandler) UpdateServer(c *gin.Context) {
	var req UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	var server models.Server
	if err := h.DB.First(&server, c.Param("id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "server not found")
		return
	}

	var existing models.Server
	err := h.DB.Where("host = ? AND id <> ?", req.Host, server.ID).First(&existing).Error
	if err == nil {
		utils.Error(c, http.StatusConflict, "server with this host already exists")
		return
	}

	server.Name = req.Name
	server.Host = req.Host
	server.OS = req.OS
	server.Description = req.Description

	if err := h.DB.Save(&server).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to update server")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"id":           server.ID,
		"name":         server.Name,
		"host":         server.Host,
		"os":           server.OS,
		"description":  server.Description,
		"is_active":    server.IsActive,
		"status":       server.Status,
		"last_seen_at": server.LastSeenAt,
		"created_at":   server.CreatedAt,
	})
}

func (h *ServerHandler) DeactivateServer(c *gin.Context) {
	var server models.Server
	if err := h.DB.First(&server, c.Param("id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "server not found")
		return
	}

	server.IsActive = false
	server.Status = "offline"

	if err := h.DB.Save(&server).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to deactivate server")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"message": "server deactivated successfully",
	})
}

func (h *ServerHandler) ActivateServer(c *gin.Context) {
	var server models.Server
	if err := h.DB.First(&server, c.Param("id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "server not found")
		return
	}

	server.IsActive = true
	if server.Status == "offline" {
		server.Status = "normal"
	}

	if err := h.DB.Save(&server).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to activate server")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"message": "server activated successfully",
	})
}

func (h *ServerHandler) GetCurrentMetrics(c *gin.Context) {
	var server models.Server
	if err := h.DB.First(&server, c.Param("id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "server not found")
		return
	}

	var metric models.CurrentMetric
	if err := h.DB.Where("server_id = ?", server.ID).First(&metric).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "current metrics not found")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"server_id":    metric.ServerID,
		"cpu_usage":    metric.CPUUsage,
		"ram_usage":    metric.RAMUsage,
		"disk_usage":   metric.DiskUsage,
		"collected_at": metric.CollectedAt,
	})
}

func (h *ServerHandler) GetMetricHistory(c *gin.Context) {
	serverID := c.Param("id")
	metricType := c.Query("metric")
	period := c.DefaultQuery("period", "1h")

	if metricType != "cpu" && metricType != "ram" && metricType != "disk" {
		utils.Error(c, http.StatusBadRequest, "metric must be cpu, ram or disk")
		return
	}

	var duration time.Duration
	switch period {
	case "1h":
		duration = time.Hour
	case "6h":
		duration = 6 * time.Hour
	case "24h":
		duration = 24 * time.Hour
	default:
		utils.Error(c, http.StatusBadRequest, "period must be 1h, 6h or 24h")
		return
	}

	var server models.Server
	if err := h.DB.First(&server, serverID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "server not found")
		return
	}

	fromTime := time.Now().Add(-duration)

	var history []models.MetricHistory
	if err := h.DB.
		Where("server_id = ? AND metric_type = ? AND collected_at >= ?", server.ID, metricType, fromTime).
		Order("collected_at asc").
		Find(&history).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to fetch metric history")
		return
	}

	response := make([]gin.H, 0, len(history))
	for _, item := range history {
		response = append(response, gin.H{
			"id":           item.ID,
			"server_id":    item.ServerID,
			"metric_type":  item.MetricType,
			"metric_value": item.MetricValue,
			"collected_at": item.CollectedAt,
		})
	}

	utils.Success(c, http.StatusOK, gin.H{
		"server_id": server.ID,
		"metric":    metricType,
		"period":    period,
		"items":     response,
	})
}
