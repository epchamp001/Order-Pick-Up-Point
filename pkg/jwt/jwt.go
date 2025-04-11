package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type TokenService interface {
	GenerateToken(userID string, role string) (string, error)
	ParseJWTToken(tokenString string) (*CustomClaims, error)
}

type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenServiceImpl struct {
	secretKey           string
	tokenExpirationTime int
}

func NewTokenService(secretKey string, tokenExpTime int) TokenService {
	return &TokenServiceImpl{secretKey: secretKey, tokenExpirationTime: tokenExpTime}
}

func (t *TokenServiceImpl) GenerateToken(userID string, role string) (string, error) {
	now := time.Now()

	expiration := now.Add(time.Duration(t.tokenExpirationTime) * time.Second)

	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(t.secretKey))
}

func (t *TokenServiceImpl) ParseJWTToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
