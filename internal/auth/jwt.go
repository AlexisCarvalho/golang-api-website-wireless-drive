package auth

import (
	"log"
	"time"
	"wireless_drive/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = mustLoadJWTSecret()

// mustLoadJWTSecret loads JWT_SECRET from the environment.
func mustLoadJWTSecret() []byte {
	secret := config.GetEnv("JWT_SECRET", "")
	if secret == "" {
		log.Fatal("JWT_SECRET is not configured")
	}
	return []byte(secret)
}

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID uint) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return token.Claims.(*Claims), nil
}
