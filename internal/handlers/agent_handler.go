package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"sermon-backend/internal/models"
	"sermon-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AgentHandler struct {
	DB *gorm.DB
}

type AgentMetricsRequest struct {
	CPUUsage    float64   `json:"cpu_usage" binding:"required"`
	RAMUsage    float64   `json:"ram_usage" binding:"required"`
	DiskUsage   float64   `json:"disk_usage" binding:"required"`
	CollectedAt time.Time `json:"collected_at" binding:"required"`
}

func NewAgentHandler(db *gorm.DB) *AgentHandler {
	return &AgentHandler{DB: db}
}

func (h *AgentHandler) ReceiveMetrics(c *gin.Context) {
	token, ok := extractBearerToken(c.GetHeader("Authorization"))
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "missing or invalid agent token")
		return
	}

	var servers []models.Server
	if err := h.DB.Where("is_active = ?", true).Find(&servers).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to load servers")
		return
	}

	var server *models.Server
	for i := range servers {
		if servers[i].AgentTokenHash != "" && utils.CheckAgentToken(token, servers[i].AgentTokenHash) {
			server = &servers[i]
			break
		}
	}

	if server == nil {
		utils.Error(c, http.StatusUnauthorized, "invalid agent token")
		return
	}

	var req AgentMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		currentMetric := models.CurrentMetric{
			ServerID:    server.ID,
			CPUUsage:    req.CPUUsage,
			RAMUsage:    req.RAMUsage,
			DiskUsage:   req.DiskUsage,
			CollectedAt: req.CollectedAt,
		}

		var existing models.CurrentMetric
		if err := tx.Where("server_id = ?", server.ID).First(&existing).Error; err == nil {
			existing.CPUUsage = req.CPUUsage
			existing.RAMUsage = req.RAMUsage
			existing.DiskUsage = req.DiskUsage
			existing.CollectedAt = req.CollectedAt
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(&currentMetric).Error; err != nil {
				return err
			}
		}

		history := []models.MetricHistory{
			{
				ServerID:    server.ID,
				MetricType:  "cpu",
				MetricValue: req.CPUUsage,
				CollectedAt: req.CollectedAt,
			},
			{
				ServerID:    server.ID,
				MetricType:  "ram",
				MetricValue: req.RAMUsage,
				CollectedAt: req.CollectedAt,
			},
			{
				ServerID:    server.ID,
				MetricType:  "disk",
				MetricValue: req.DiskUsage,
				CollectedAt: req.CollectedAt,
			},
		}

		if err := tx.Create(&history).Error; err != nil {
			return err
		}

		status := "normal"

		if err := h.processMetric(tx, server.ID, "cpu", req.CPUUsage, &status, req.CollectedAt); err != nil {
			return err
		}
		if err := h.processMetric(tx, server.ID, "ram", req.RAMUsage, &status, req.CollectedAt); err != nil {
			return err
		}
		if err := h.processMetric(tx, server.ID, "disk", req.DiskUsage, &status, req.CollectedAt); err != nil {
			return err
		}

		server.Status = status
		unix := req.CollectedAt.Unix()
		server.LastSeenAt = &unix

		if err := tx.Save(server).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to process metrics")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"message":   "metrics processed",
		"server_id": server.ID,
	})
}

func (h *AgentHandler) processMetric(tx *gorm.DB, serverID uint, metricType string, actualValue float64, serverStatus *string, collectedAt time.Time) error {
	var threshold models.Threshold
	if err := tx.Where("server_id = ? AND metric_type = ?", serverID, metricType).First(&threshold).Error; err != nil {
		return err
	}

	if actualValue >= float64(threshold.CriticalValue) {
		*serverStatus = "critical"
		return h.openOrUpdateIncident(tx, serverID, metricType, float64(threshold.CriticalValue), actualValue, collectedAt)
	}

	if actualValue >= float64(threshold.WarningValue) {
		if *serverStatus != "critical" {
			*serverStatus = "warning"
		}
		return h.openOrUpdateIncident(tx, serverID, metricType, float64(threshold.WarningValue), actualValue, collectedAt)
	}

	return nil
}

func (h *AgentHandler) openOrUpdateIncident(tx *gorm.DB, serverID uint, metricType string, thresholdValue, actualValue float64, startedAt time.Time) error {
	var incident models.Incident
	err := tx.Where("server_id = ? AND metric_type = ? AND status IN ?", serverID, metricType, []string{"open", "in_progress"}).
		First(&incident).Error

	if err == nil {
		incident.ActualValue = actualValue
		incident.ThresholdValue = thresholdValue
		incident.Message = fmt.Sprintf("%s threshold exceeded: actual %.2f, threshold %.2f", metricType, actualValue, thresholdValue)
		return tx.Save(&incident).Error
	}

	if err != gorm.ErrRecordNotFound {
		return err
	}

	newIncident := models.Incident{
		ServerID:       serverID,
		MetricType:     metricType,
		Status:         "open",
		ThresholdValue: thresholdValue,
		ActualValue:    actualValue,
		Message:        fmt.Sprintf("%s threshold exceeded: actual %.2f, threshold %.2f", metricType, actualValue, thresholdValue),
		StartedAt:      startedAt,
	}

	return tx.Create(&newIncident).Error
}

func extractBearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", false
	}

	return parts[1], true
}
