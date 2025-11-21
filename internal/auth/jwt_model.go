package auth

type JWTResponse struct {
    AccessToken        string
    RefreshToken       string
    AccessExpiresIn    int64
    RefreshExpiresIn   int64
    TokenType          string
}

type JWTPayload struct {
    UserID string
}