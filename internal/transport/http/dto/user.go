package dto

// UserLoginRequest is the request body for user login.
type UserLoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserRegistrationRequest is the request body for user registration.
type UserRegistrationRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	Telegram string `json:"telegram"`
}

// TokenRefreshRequest is the request body for refreshing tokens.
type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UserResponse is the response body for a user.
type UserResponse struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Email    string `json:"email"`
	Telegram string `json:"telegram"`
}

// JWTResponse is the response body containing JWT tokens.
type JWTResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
