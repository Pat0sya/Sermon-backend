package handlers

import (
	"net/http"

	"sermon-backend/internal/models"
	"sermon-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	DB *gorm.DB
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	var users []models.User
	if err := h.DB.Preload("Role").Order("id asc").Find(&users).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	response := make([]gin.H, 0, len(users))
	for _, user := range users {
		response = append(response, gin.H{
			"id":                   user.ID,
			"username":             user.Username,
			"role":                 user.Role.Name,
			"must_change_password": user.MustChangePassword,
			"is_active":            user.IsActive,
			"created_at":           user.CreatedAt,
		})
	}

	utils.Success(c, http.StatusOK, response)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Username) < 3 {
		utils.Error(c, http.StatusBadRequest, "username must be at least 3 characters")
		return
	}

	if len(req.Password) < 6 {
		utils.Error(c, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	if req.Role != "operator" && req.Role != "admin" {
		utils.Error(c, http.StatusBadRequest, "role must be admin or operator")
		return
	}

	var existing models.User
	if err := h.DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		utils.Error(c, http.StatusConflict, "user already exists")
		return
	}

	var role models.Role
	if err := h.DB.Where("name = ?", req.Role).First(&role).Error; err != nil {
		utils.Error(c, http.StatusBadRequest, "role not found")
		return
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := models.User{
		Username:           req.Username,
		PasswordHash:       passwordHash,
		RoleID:             role.ID,
		MustChangePassword: true,
		IsActive:           true,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	if err := h.DB.Preload("Role").First(&user, user.ID).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to load created user")
		return
	}

	utils.Success(c, http.StatusCreated, gin.H{
		"id":                   user.ID,
		"username":             user.Username,
		"role":                 user.Role.Name,
		"must_change_password": user.MustChangePassword,
		"is_active":            user.IsActive,
		"created_at":           user.CreatedAt,
	})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.NewPassword) < 6 {
		utils.Error(c, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	var user models.User
	if err := h.DB.Preload("Role").First(&user, c.Param("id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "user not found")
		return
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user.PasswordHash = newHash
	user.MustChangePassword = true

	if err := h.DB.Save(&user).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to reset password")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"message": "password reset successfully",
		"user": gin.H{
			"id":                   user.ID,
			"username":             user.Username,
			"role":                 user.Role.Name,
			"must_change_password": user.MustChangePassword,
			"is_active":            user.IsActive,
		},
	})
}

func (h *UserHandler) DeactivateUser(c *gin.Context) {
	targetUserID := c.Param("id")
	currentUserID := c.GetUint("user_id")

	var user models.User
	if err := h.DB.Preload("Role").First(&user, targetUserID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "user not found")
		return
	}

	if user.ID == currentUserID {
		utils.Error(c, http.StatusBadRequest, "you cannot deactivate yourself")
		return
	}

	user.IsActive = false

	if err := h.DB.Save(&user).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to deactivate user")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"message": "user deactivated successfully",
		"user": gin.H{
			"id":                   user.ID,
			"username":             user.Username,
			"role":                 user.Role.Name,
			"must_change_password": user.MustChangePassword,
			"is_active":            user.IsActive,
		},
	})
}

func (h *UserHandler) ActivateUser(c *gin.Context) {
	var user models.User
	if err := h.DB.Preload("Role").First(&user, c.Param("id")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "user not found")
		return
	}

	user.IsActive = true

	if err := h.DB.Save(&user).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to activate user")
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"message": "user activated successfully",
		"user": gin.H{
			"id":                   user.ID,
			"username":             user.Username,
			"role":                 user.Role.Name,
			"must_change_password": user.MustChangePassword,
			"is_active":            user.IsActive,
		},
	})
}
