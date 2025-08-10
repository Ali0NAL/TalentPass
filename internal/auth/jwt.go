package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func mustEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func GenerateAccessToken(userID int64, email string) (string, time.Time, error) {
	secret := []byte(mustEnv("JWT_SECRET", "dev-secret-change-me"))
	ttl := mustEnv("ACCESS_TOKEN_TTL", "15m")
	dur, err := time.ParseDuration(ttl)
	if err != nil {
		dur = 15 * time.Minute
	}
	exp := time.Now().Add(dur)

	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(secret)
	return signed, exp, err
}

func Parse(tokenStr string) (*Claims, error) {
	secret := []byte(mustEnv("JWT_SECRET", "dev-secret-change-me"))
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if c, ok := tok.Claims.(*Claims); ok && tok.Valid {
		return c, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
