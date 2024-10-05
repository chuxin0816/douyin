package jwt

import (
	"time"

	"douyin/src/config"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

const (
	issuer        = "chuxin"
	tokenDuration = time.Hour * 24 * 30
)

type Claims struct {
	UserID int64
	*jwt.RegisteredClaims
}

func Init() {
	jwtKey = []byte(config.Conf.JwtKey)
}

func GenerateToken(userID int64) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenDuration)),
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtKey)
}

func ParseToken(tokenStr string) (userID *int64) {
	if tokenStr == "" {
		return nil
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil
	}

	return &claims.UserID
}
