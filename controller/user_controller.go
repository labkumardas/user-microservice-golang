package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	apperrors "user-microservice-golang/errors"
	"user-microservice-golang/interfaces"
	"user-microservice-golang/middleware"
	"user-microservice-golang/model"
	"user-microservice-golang/utils"
)

// UserController handles HTTP requests and delegates to the service layer
type UserController struct {
	service interfaces.UserService
	logger  *zap.Logger
}

// NewUserController constructs a UserController
func NewUserController(svc interfaces.UserService, logger *zap.Logger) *UserController {
	return &UserController{service: svc, logger: logger}
}

// ─── Auth endpoints ───────────────────────────────────────────────────────────

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body model.RegisterRequest true "Registration payload"
// @Success      201  {object} utils.APIResponse{data=model.AuthResponse}
// @Failure      400  {object} utils.APIResponse
// @Failure      409  {object} utils.APIResponse
// @Router       /auth/register [post]

func (ctrl *UserController) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	fmt.Println(&req)

	result, err := ctrl.service.Register(c.Request.Context(), &req)
	if err != nil {
		ae := apperrors.Map(err)
		utils.RespondError(c, ae.HTTPCode, ae.Message)
		return
	}

	utils.RespondCreated(c, result)
}

// Login godoc
// @Summary      Authenticate and get a JWT
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body model.LoginRequest true "Login credentials"
// @Success      200  {object} utils.APIResponse{data=model.AuthResponse}
// @Failure      401  {object} utils.APIResponse
// @Router       /auth/login [post]

func (ctrl *UserController) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := ctrl.service.Login(c.Request.Context(), &req)
	if err != nil {
		ae := apperrors.Map(err)
		utils.RespondError(c, ae.HTTPCode, ae.Message)
		return
	}

	utils.RespondSuccess(c, result)
}

// ─── User endpoints ───────────────────────────────────────────────────────────

// GetMe godoc
// @Summary      Get the currently authenticated user
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} utils.APIResponse{data=model.UserResponse}
// @Router       /users/me [get]

func (ctrl *UserController) GetMe(c *gin.Context) {
	id, _ := c.Get(middleware.AuthUserIDKey)
	result, err := ctrl.service.GetByID(c.Request.Context(), id.(string))
	if err != nil {
		ae := apperrors.Map(err)
		utils.RespondError(c, ae.HTTPCode, ae.Message)
		return
	}
	utils.RespondSuccess(c, result)
}

// GetUserByID godoc
// @Summary      Get a user by ID (admin or self)
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} utils.APIResponse{data=model.UserResponse}
// @Failure      404 {object} utils.APIResponse
// @Router       /users/{id} [get]

func (ctrl *UserController) GetUserByID(c *gin.Context) {
	result, err := ctrl.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		ae := apperrors.Map(err)
		utils.RespondError(c, ae.HTTPCode, ae.Message)
		return
	}
	utils.RespondSuccess(c, result)
}

// GetAllUsers godoc
// @Summary      List all users with pagination (admin only)
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        page  query int false "Page number"  default(1)
// @Param        limit query int false "Page size"    default(10)
// @Success      200 {object} utils.APIResponse{data=model.PaginatedUsersResponse}
// @Router       /users [get]

func (ctrl *UserController) GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := ctrl.service.GetAll(c.Request.Context(), page, limit)
	if err != nil {
		ae := apperrors.Map(err)
		utils.RespondError(c, ae.HTTPCode, ae.Message)
		return
	}
	utils.RespondSuccess(c, result)
}

// UpdateProfile godoc
// @Summary      Update the authenticated user's profile
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string                    true "User ID"
// @Param        body body model.UpdateUserRequest   true "Update payload"
// @Success      200 {object} utils.APIResponse{data=model.UserResponse}
// @Router       /users/{id} [patch]

func (ctrl *UserController) UpdateProfile(c *gin.Context) {
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := ctrl.service.UpdateProfile(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		ae := apperrors.Map(err)
		utils.RespondError(c, ae.HTTPCode, ae.Message)
		return
	}
	utils.RespondSuccess(c, result)
}

// UpdatePassword godoc
// @Summary      Change the authenticated user's password
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string                       true "User ID"
// @Param        body body model.UpdatePasswordRequest  true "Password payload"
// @Success      200 {object} utils.APIResponse
// @Router       /users/{id}/password [patch]
func (ctrl *UserController) UpdatePassword(c *gin.Context) {
	var req model.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := ctrl.service.UpdatePassword(c.Request.Context(), c.Param("id"), &req); err != nil {
		ae := apperrors.Map(err)
		utils.RespondError(c, ae.HTTPCode, ae.Message)
		return
	}
	utils.RespondMessage(c, http.StatusOK, "password updated successfully")
}

// UpdateStatus godoc
// @Summary      Change a user's account status (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string                     true "User ID"
// @Param        body body model.UpdateStatusRequest  true "Status payload"
// @Success      200 {object} utils.APIResponse{data=model.UserResponse}
// @Router       /users/{id}/status [patch]
func (ctrl *UserController) UpdateStatus(c *gin.Context) {
	var req model.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := ctrl.service.UpdateStatus(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		ae := apperrors.Map(err)
		utils.RespondError(c, ae.HTTPCode, ae.Message)
		return
	}
	utils.RespondSuccess(c, result)
}

// DeleteUser godoc
// @Summary      Soft-delete a user (admin or self)
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} utils.APIResponse
// @Router       /users/{id} [delete]
func (ctrl *UserController) DeleteUser(c *gin.Context) {
	if err := ctrl.service.DeleteUser(c.Request.Context(), c.Param("id")); err != nil {
		ae := apperrors.Map(err)
		utils.RespondError(c, ae.HTTPCode, ae.Message)
		return
	}
	utils.RespondMessage(c, http.StatusOK, "user deleted successfully")
}

// HealthCheck godoc
// @Summary      Liveness probe
// @Tags         system
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /health [get]
func (ctrl *UserController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "user-microservice-golang"})
}
