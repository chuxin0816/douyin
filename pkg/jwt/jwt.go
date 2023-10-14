package jwt

import (
	"douyin/config"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	jwtKey          = []byte(config.Conf.JwtKey)
	ErrInvalidToken = errors.New("invalid token")
)

const (
	issuer        = "chuxin"
	tokenDuration = time.Hour * 24 * 30
)

type Claims struct {
	UserID int64
	*jwt.StandardClaims
}

func GenerateToken(userID int64) (string, error) {
	claims := &Claims{
		UserID: userID,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenDuration).Unix(),
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ParseToken(tokenStr string) (userID int64, err error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return 0, err
	}
	if !token.Valid {
		return 0, ErrInvalidToken
	}
	return claims.UserID, nil
}
