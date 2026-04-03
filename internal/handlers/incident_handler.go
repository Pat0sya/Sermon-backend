package handlers

import (
	"net/http"
	"time"

	"sermon-backend/internal/models"
	"sermon-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IncidentHandler struct {
	DB *gorm.DB
}

type UpdateIncidentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type AddCommentRequest struct {
	Comment string `json:"comment" binding:"required"`
}

func NewIncidentHandler(db *gorm.DB) *IncidentHandler {
	return &IncidentHandler{DB: db}
}

func (h *IncidentHandler) GetIncidents(c *gin.Context) {
	status := c.Query("status")

	query := h.DB.Preload("Server").Order("created_at desc")

	if status != "" {
		if status != "open" && status != "in_progress" && status != "closed" {
			utils.Error(c, http.StatusBadRequest, "invalid status")
			return
		}
		query = query.Where("status = ?", status)
	}

	var incidents []models.Incident
	if err := query.Find(&incidents).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to fetch incidents")
		return
	}

	response := make([]gin.H, 0, len(incidents))
	for _, i := range incidents {
		response = append(response, gin.H{
			"id":        i.ID,
			"server_id": i.ServerID,
			"server": gin.H{
				"id":           i.Server.ID,
				"name":         i.Server.Name,
				"host":         i.Server.Host,
				"os":           i.Server.OS,
				"is_active":    i.Server.IsActive,
				"status":       i.Server.Status,
				"last_seen_at": i.Server.LastSeenAt,
			},
			"metric_type":     i.MetricType,
			"status":          i.Status,
			"threshold_value": i.ThresholdValue,
			"actual_value":    i.ActualValue,
			"message":         i.Message,
			"started_at":      i.StartedAt,
			"closed_at":       i.ClosedAt,
		})
	}

	utils.Success(c, http.StatusOK, response)
}

func (h *IncidentHandler) GetIncidentByID(c *gin.Context) {
	var incident models.Incident
	if err := h.DB.Preload("Server").First(&incident, c.Param("id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "incident not found")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"id":        incident.ID,
		"server_id": incident.ServerID,
		"server": gin.H{
			"id":           incident.Server.ID,
			"name":         incident.Server.Name,
			"host":         incident.Server.Host,
			"os":           incident.Server.OS,
			"is_active":    incident.Server.IsActive,
			"status":       incident.Server.Status,
			"last_seen_at": incident.Server.LastSeenAt,
		},
		"metric_type":     incident.MetricType,
		"status":          incident.Status,
		"threshold_value": incident.ThresholdValue,
		"actual_value":    incident.ActualValue,
		"message":         incident.Message,
		"started_at":      incident.StartedAt,
		"closed_at":       incident.ClosedAt,
	})
}

func (h *IncidentHandler) UpdateIncidentStatus(c *gin.Context) {
	var req UpdateIncidentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Status != "open" && req.Status != "in_progress" && req.Status != "closed" {
		utils.Error(c, http.StatusBadRequest, "invalid status")
		return
	}

	role := c.GetString("role")
	if req.Status == "closed" && role != "admin" {
		utils.Error(c, http.StatusForbidden, "only admin can close incidents")
		return
	}

	var incident models.Incident
	if err := h.DB.First(&incident, c.Param("id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "incident not found")
		return
	}

	incident.Status = req.Status
	if req.Status == "closed" {
		now := time.Now()
		incident.ClosedAt = &now
	} else {
		incident.ClosedAt = nil
	}

	if err := h.DB.Save(&incident).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to update incident")
		return
	}

	if err := h.DB.Preload("Server").First(&incident, incident.ID).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to fetch updated incident")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"id":        incident.ID,
		"server_id": incident.ServerID,
		"server": gin.H{
			"id":           incident.Server.ID,
			"name":         incident.Server.Name,
			"host":         incident.Server.Host,
			"os":           incident.Server.OS,
			"is_active":    incident.Server.IsActive,
			"status":       incident.Server.Status,
			"last_seen_at": incident.Server.LastSeenAt,
		},
		"metric_type":     incident.MetricType,
		"status":          incident.Status,
		"threshold_value": incident.ThresholdValue,
		"actual_value":    incident.ActualValue,
		"message":         incident.Message,
		"started_at":      incident.StartedAt,
		"closed_at":       incident.ClosedAt,
	})
}

func (h *IncidentHandler) GetComments(c *gin.Context) {
	var comments []models.IncidentComment
	if err := h.DB.Preload("User").Preload("User.Role").Where("incident_id = ?", c.Param("id")).Order("created_at asc").Find(&comments).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to fetch comments")
		return
	}

	response := make([]gin.H, 0, len(comments))
	for _, comment := range comments {
		response = append(response, gin.H{
			"id":          comment.ID,
			"incident_id": comment.IncidentID,
			"user_id":     comment.UserID,
			"comment":     comment.Comment,
			"created_at":  comment.CreatedAt,
			"user": gin.H{
				"id":       comment.User.ID,
				"username": comment.User.Username,
				"role":     comment.User.Role.Name,
			},
		})
	}

	utils.Success(c, http.StatusOK, response)
}

func (h *IncidentHandler) AddComment(c *gin.Context) {
	var req AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	var incident models.Incident
	if err := h.DB.First(&incident, c.Param("id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "incident not found")
		return
	}

	comment := models.IncidentComment{
		IncidentID: incident.ID,
		UserID:     c.GetUint("user_id"),
		Comment:    req.Comment,
	}

	if err := h.DB.Create(&comment).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to add comment")
		return
	}

	if err := h.DB.Preload("User").Preload("User.Role").First(&comment, comment.ID).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to fetch created comment")
		return
	}

	utils.Success(c, http.StatusCreated, gin.H{
		"id":          comment.ID,
		"incident_id": comment.IncidentID,
		"user_id":     comment.UserID,
		"comment":     comment.Comment,
		"created_at":  comment.CreatedAt,
		"user": gin.H{
			"id":       comment.User.ID,
			"username": comment.User.Username,
			"role":     comment.User.Role.Name,
		},
	})
}
