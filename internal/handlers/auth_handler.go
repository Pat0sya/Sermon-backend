package handlers

import (
	"net/http"

	"sermon-backend/internal/config"
	"sermon-backend/internal/models"
	"sermon-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{DB: db, Cfg: cfg}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	var user models.User
	err := h.DB.Preload("Role").Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if !user.IsActive {
		utils.Error(c, http.StatusForbidden, "user is inactive")
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		utils.Error(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := utils.GenerateJWT(h.Cfg.JWTSecret, user.ID, user.Username, user.Role.Name)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":                   user.ID,
			"username":             user.Username,
			"role":                 user.Role.Name,
			"must_change_password": user.MustChangePassword,
			"is_active":            user.IsActive,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	var user models.User
	if err := h.DB.Preload("Role").First(&user, c.GetUint("user_id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "user not found")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"user": gin.H{
			"id":                   user.ID,
			"username":             user.Username,
			"role":                 user.Role.Name,
			"must_change_password": user.MustChangePassword,
			"is_active":            user.IsActive,
		},
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.NewPassword) < 6 {
		utils.Error(c, http.StatusBadRequest, "new password must be at least 6 characters")
		return
	}

	var user models.User
	if err := h.DB.First(&user, c.GetUint("user_id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "user not found")
		return
	}

	if !utils.CheckPasswordHash(req.OldPassword, user.PasswordHash) {
		utils.Error(c, http.StatusUnauthorized, "old password is incorrect")
		return
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user.PasswordHash = newHash
	user.MustChangePassword = false

	if err := h.DB.Save(&user).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to update password")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"message": "password changed successfully",
	})
}
