package jwt

import (
	"douyin/config"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

var jwtKey = []byte(config.Conf.JwtKey)
var ErrInvalidToken = errors.New("invalid token")

const issuer = "chuxin"
const tokenDuration = time.Hour * 24

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

func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
