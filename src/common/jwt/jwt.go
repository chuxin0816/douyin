package jwt

import (
	"time"

	"douyin/src/config"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

const (
	issuer               = "chuxin"
	accessTokenDuration  = time.Hour * 1
	refreshTokenDuration = time.Hour * 24 * 30
)

type Claims struct {
	UserID  int64
	Refresh bool
	*jwt.RegisteredClaims
}

func Init() {
	jwtKey = []byte(config.Conf.JwtKey)
}

func GenerateAccessToken(userID int64) (string, error) {
	claims := &Claims{
		UserID:  userID,
		Refresh: false,
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenDuration)),
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtKey)
}

func GenerateRefreshToken(userID int64) (string, error) {
	claims := &Claims{
		UserID:  userID,
		Refresh: true,
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenDuration)),
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtKey)
}

func ParseAccessToken(tokenStr string) (userID *int64) {
	if tokenStr == "" {
		return nil
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid || claims.Refresh {
		return nil
	}

	return &claims.UserID
}

func ParseRefreshToken(tokenStr string) (userID *int64) {
	if tokenStr == "" {
		return nil
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid || !claims.Refresh {
		return nil
	}

	return &claims.UserID
}
