package jwt

import (
	"time"

	"github.com/chuxin0816/Scaffold/config"
	"github.com/golang-jwt/jwt"
)

var jwtKey = []byte(config.Conf.JwtKey)

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
			Issuer:    "chuxin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
