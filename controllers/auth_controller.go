package controllers

import (
	"errors"
	"log"
	"net/http"

	"go-ecommerce/config"
	"go-ecommerce/internal/auth"
	"go-ecommerce/middleware"
	"go-ecommerce/models"
	"go-ecommerce/services"

	"github.com/gin-gonic/gin"
)

// AuthController handles authentication-related HTTP requests
type AuthController struct {
	service *services.AuthService
}

// NewAuthController creates a new auth controller. It returns an error so that
// invalid auth configuration stops the process at startup instead of failing on
// the first login.
func NewAuthController(cfg config.AuthConfig) (*AuthController, error) {
	service, err := services.NewAuthService(cfg)
	if err != nil {
		return nil, err
	}
	return &AuthController{service: service}, nil
}

// Service exposes the auth service so routes can wire up the middleware with
// the same TokenManager that issues tokens.
func (c *AuthController) Service() *services.AuthService {
	return c.service
}

// Signup handles POST /api/auth/signup
// @Summary      Register a new account
// @Description  Creates a customer account and returns an access/refresh token pair. Passwords must be at least 8 characters, contain a letter plus a number or symbol, and must not echo your name or email.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.SignupRequest  true  "Registration details"
// @Success      201      {object}  models.AuthResponse
// @Failure      400      {object}  models.ErrorResponse  "Validation or password policy failure"
// @Failure      409      {object}  models.ErrorResponse  "Email already registered"
// @Failure      429      {object}  models.ErrorResponse  "Too many requests"
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/signup [post]
func (c *AuthController) Signup(ctx *gin.Context) {
	var req models.SignupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	response, err := c.service.Signup(&req, ctx.Request.UserAgent(), ctx.ClientIP())
	if err != nil {
		respondAuthError(ctx, "signup", err)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// Login handles POST /api/auth/login
// @Summary      Log in
// @Description  Verifies credentials and returns an access/refresh token pair. Repeated failures lock the account temporarily.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.LoginRequest  true  "Credentials"
// @Success      200      {object}  models.AuthResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse  "Invalid email or password"
// @Failure      403      {object}  models.ErrorResponse  "Account inactive"
// @Failure      429      {object}  models.ErrorResponse  "Account locked or rate limited"
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req models.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	response, err := c.service.Login(&req, ctx.Request.UserAgent(), ctx.ClientIP())
	if err != nil {
		respondAuthError(ctx, "login", err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// Refresh handles POST /api/auth/refresh
// @Summary      Refresh the access token
// @Description  Exchanges a refresh token for a new pair. The old refresh token is invalidated on use, so always store the one returned here.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.RefreshRequest  true  "Refresh token"
// @Success      200      {object}  models.AuthResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse  "Invalid, expired or already-used refresh token"
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/refresh [post]
func (c *AuthController) Refresh(ctx *gin.Context) {
	var req models.RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	response, err := c.service.Refresh(req.RefreshToken, ctx.Request.UserAgent(), ctx.ClientIP())
	if err != nil {
		respondAuthError(ctx, "refresh", err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// Logout handles POST /api/auth/logout
// @Summary      Log out of this session
// @Description  Revokes the supplied refresh token. Idempotent: an unknown token also returns success. The access token remains valid until it expires, so clients should discard it.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.RefreshRequest  true  "Refresh token to revoke"
// @Success      200      {object}  models.MessageResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	var req models.RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := c.service.Logout(req.RefreshToken); err != nil {
		respondAuthError(ctx, "logout", err)
		return
	}

	ctx.JSON(http.StatusOK, models.MessageResponse{Message: "logged out successfully"})
}

// LogoutAll handles POST /api/auth/logout-all
// @Summary      Log out everywhere
// @Description  Revokes every refresh token belonging to the authenticated user.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.MessageResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /auth/logout-all [post]
func (c *AuthController) LogoutAll(ctx *gin.Context) {
	userID, ok := middleware.UserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "authentication required"})
		return
	}

	if err := c.service.LogoutAll(userID); err != nil {
		respondAuthError(ctx, "logout-all", err)
		return
	}

	ctx.JSON(http.StatusOK, models.MessageResponse{Message: "all sessions revoked successfully"})
}

// Me handles GET /api/auth/me
// @Summary      Get the current user
// @Description  Returns the profile of the authenticated user. The identity comes from the token, never from a request parameter.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.UserResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /auth/me [get]
func (c *AuthController) Me(ctx *gin.Context) {
	userID, ok := middleware.UserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "authentication required"})
		return
	}

	user, err := c.service.GetProfile(userID)
	if err != nil {
		respondAuthError(ctx, "profile", err)
		return
	}

	ctx.JSON(http.StatusOK, models.NewUserResponse(user))
}

// ChangePassword handles POST /api/auth/change-password
// @Summary      Change the password
// @Description  Requires the current password. On success every other session is signed out.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      models.ChangePasswordRequest  true  "Current and new password"
// @Success      200      {object}  models.MessageResponse
// @Failure      400      {object}  models.ErrorResponse  "Validation or password policy failure"
// @Failure      401      {object}  models.ErrorResponse  "Current password incorrect"
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/change-password [post]
func (c *AuthController) ChangePassword(ctx *gin.Context) {
	userID, ok := middleware.UserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "authentication required"})
		return
	}

	var req models.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := c.service.ChangePassword(userID, &req); err != nil {
		respondAuthError(ctx, "change-password", err)
		return
	}

	ctx.JSON(http.StatusOK, models.MessageResponse{Message: "password changed successfully, other sessions were signed out"})
}

// respondAuthError translates a service error into an HTTP response.
//
// Only known sentinel errors reach the client. Anything unexpected is logged
// server-side and answered with a generic message, because raw error strings
// leak schema details, query fragments and file paths to an attacker.
func respondAuthError(ctx *gin.Context, operation string, err error) {
	switch {
	case errors.Is(err, services.ErrEmailTaken):
		ctx.JSON(http.StatusConflict, models.ErrorResponse{Error: err.Error()})

	case errors.Is(err, services.ErrInvalidCredentials),
		errors.Is(err, services.ErrInvalidRefreshToken):
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: err.Error()})

	case errors.Is(err, services.ErrAccountLocked):
		ctx.JSON(http.StatusTooManyRequests, models.ErrorResponse{Error: err.Error()})

	case errors.Is(err, services.ErrAccountInactive):
		ctx.JSON(http.StatusForbidden, models.ErrorResponse{Error: err.Error()})

	case errors.Is(err, services.ErrUserNotFound):
		ctx.JSON(http.StatusNotFound, models.ErrorResponse{Error: err.Error()})

	// Password policy failures are safe to echo verbatim: the user needs to
	// know exactly which rule they broke.
	case errors.Is(err, services.ErrSamePassword),
		errors.Is(err, auth.ErrPasswordTooShort),
		errors.Is(err, auth.ErrPasswordTooLong),
		errors.Is(err, auth.ErrPasswordTooCommon),
		errors.Is(err, auth.ErrPasswordTooSimple),
		errors.Is(err, auth.ErrPasswordEchoesID):
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})

	default:
		log.Printf("auth %s failed: %v", operation, err)
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "an unexpected error occurred, please try again",
		})
	}
}
