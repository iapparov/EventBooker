package handlers

import (
	"eventbooker/internal/auth"
	"eventbooker/internal/domain/user"
	wbgin "github.com/wb-go/wbf/ginext"
	"eventbooker/internal/web/dto"
	"net/http"
)

type UserHandler struct {
	Service UserIFace
}

type UserIFace interface {
	Login(Login, Password string) (*auth.JWTResponse, error)
	Registration(Login, Password, Email, Telegram string) (*user.User, error)
	RefreshTokens(tokenStr string) (*auth.JWTResponse, error)
	ValidateTokens(tokenStr string) (*auth.JWTPayload, error)
}

func NewUserHandler (service UserIFace) *UserHandler {
	return &UserHandler{
		Service: service,
	}
}


func (h *UserHandler) RegisterUser(ctx *wbgin.Context) {
	var req dto.UserRegistrationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}
	user, err := h.Service.Registration(req.Login, req.Password, req.Email, req.Telegram)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}
	res := dto.UserResponse{
		ID:       user.Id.String(),
		Login:    user.Login,
		Email:    user.Email,
		Telegram: user.Telegram,
	}
	ctx.JSON(http.StatusOK, res)
}

func (h *UserHandler) LoginUser(ctx *wbgin.Context) {
	var req dto.UserLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}
	jwtResp, err := h.Service.Login(req.Login, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, wbgin.H{"error": err.Error()})
		return
	}
	res := dto.JWTResponse{
		AccessToken:  jwtResp.AccessToken,
		RefreshToken: jwtResp.RefreshToken,
	}
	ctx.JSON(http.StatusOK, res)
}

func (h *UserHandler) RefreshToken(ctx *wbgin.Context) {
	var req dto.TokenRefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}
	jwtResp, err := h.Service.RefreshTokens(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, wbgin.H{"error": err.Error()})
		return
	}
	res := dto.JWTResponse{
		AccessToken:  jwtResp.AccessToken,
		RefreshToken: jwtResp.RefreshToken,
	}
	ctx.JSON(http.StatusOK, res)
}