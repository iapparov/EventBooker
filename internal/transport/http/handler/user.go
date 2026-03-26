package handler

import (
	"context"
	"net/http"

	"eventbooker/internal/auth"
	"eventbooker/internal/domain/user"
	"eventbooker/internal/transport/http/dto"

	wbgin "github.com/wb-go/wbf/ginext"
)

// UserServicer defines the user service interface used by UserHandler.
type UserServicer interface {
	Login(ctx context.Context, login, password string) (*auth.Response, error)
	Register(ctx context.Context, login, password, email, telegram string) (*user.User, error)
	RefreshTokens(refreshToken string) (*auth.Response, error)
	ValidateToken(tokenStr string) (*auth.Payload, error)
}

// UserHandler handles HTTP requests for user operations.
type UserHandler struct {
	service UserServicer
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(service UserServicer) *UserHandler {
	return &UserHandler{service: service}
}

// RegisterUser godoc
// @Summary      Register a new user
// @Description  Create a new user account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      dto.UserRegistrationRequest  true  "User registration info"
// @Success      200   {object}  dto.UserResponse
// @Failure      400   {object}  map[string]string  "Invalid request"
// @Failure      500   {object}  map[string]string  "Internal server error"
// @Router       /users/register [post]
func (h *UserHandler) RegisterUser(ctx *wbgin.Context) {
	var req dto.UserRegistrationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}

	u, err := h.service.Register(ctx.Request.Context(), req.Login, req.Password, req.Email, req.Telegram)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.UserResponse{
		ID:       u.ID.String(),
		Login:    u.Login,
		Email:    u.Email,
		Telegram: u.Telegram,
	})
}

// LoginUser godoc
// @Summary      Login user
// @Description  Authenticate user and return JWT tokens
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      dto.UserLoginRequest  true  "User login info"
// @Success      200   {object}  dto.JWTResponse
// @Failure      400   {object}  map[string]string  "Invalid request"
// @Failure      401   {object}  map[string]string  "Unauthorized"
// @Router       /users/login [post]
func (h *UserHandler) LoginUser(ctx *wbgin.Context) {
	var req dto.UserLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}

	jwtResp, err := h.service.Login(ctx.Request.Context(), req.Login, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, wbgin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.JWTResponse{
		AccessToken:  jwtResp.AccessToken,
		RefreshToken: jwtResp.RefreshToken,
	})
}

// RefreshToken godoc
// @Summary      Refresh JWT token
// @Description  Refresh access and refresh tokens using existing refresh token
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      dto.TokenRefreshRequest  true  "Refresh token request"
// @Success      200   {object}  dto.JWTResponse
// @Failure      400   {object}  map[string]string  "Invalid request"
// @Failure      401   {object}  map[string]string  "Unauthorized"
// @Router       /users/refresh-token [post]
func (h *UserHandler) RefreshToken(ctx *wbgin.Context) {
	var req dto.TokenRefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}

	jwtResp, err := h.service.RefreshTokens(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, wbgin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.JWTResponse{
		AccessToken:  jwtResp.AccessToken,
		RefreshToken: jwtResp.RefreshToken,
	})
}
