package handlers

import (
	"net/http"

	"sermon-backend/internal/models"
	"sermon-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ThresholdHandler struct {
	DB *gorm.DB
}

type UpdateThresholdRequest struct {
	WarningValue  int `json:"warning_value" binding:"required"`
	CriticalValue int `json:"critical_value" binding:"required"`
}

func NewThresholdHandler(db *gorm.DB) *ThresholdHandler {
	return &ThresholdHandler{DB: db}
}

func (h *ThresholdHandler) GetServerThresholds(c *gin.Context) {
	serverID := c.Param("id")

	var server models.Server
	if err := h.DB.First(&server, serverID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "server not found")
		return
	}

	var thresholds []models.Threshold
	if err := h.DB.Where("server_id = ?", server.ID).Order("id asc").Find(&thresholds).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to fetch thresholds")
		return
	}

	response := make([]gin.H, 0, len(thresholds))
	for _, t := range thresholds {
		response = append(response, gin.H{
			"id":             t.ID,
			"server_id":      t.ServerID,
			"metric_type":    t.MetricType,
			"warning_value":  t.WarningValue,
			"critical_value": t.CriticalValue,
		})
	}

	utils.Success(c, http.StatusOK, response)
}

func (h *ThresholdHandler) UpdateServerThreshold(c *gin.Context) {
	serverID := c.Param("id")
	metricType := c.Param("metric_type")

	if metricType != "cpu" && metricType != "ram" && metricType != "disk" {
		utils.Error(c, http.StatusBadRequest, "metric_type must be cpu, ram or disk")
		return
	}

	var server models.Server
	if err := h.DB.First(&server, serverID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "server not found")
		return
	}

	var req UpdateThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.WarningValue < 0 || req.WarningValue > 100 || req.CriticalValue < 0 || req.CriticalValue > 100 {
		utils.Error(c, http.StatusBadRequest, "threshold values must be between 0 and 100")
		return
	}

	if req.WarningValue >= req.CriticalValue {
		utils.Error(c, http.StatusBadRequest, "warning_value must be less than critical_value")
		return
	}

	var threshold models.Threshold
	if err := h.DB.Where("server_id = ? AND metric_type = ?", server.ID, metricType).First(&threshold).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "threshold not found")
		return
	}

	threshold.WarningValue = req.WarningValue
	threshold.CriticalValue = req.CriticalValue

	if err := h.DB.Save(&threshold).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to update threshold")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"id":             threshold.ID,
		"server_id":      threshold.ServerID,
		"metric_type":    threshold.MetricType,
		"warning_value":  threshold.WarningValue,
		"critical_value": threshold.CriticalValue,
	})
}
