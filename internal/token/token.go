package token

import (
	"fmt"
	"strings"
	"time"

	"awesomeProject/users"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

var secretKey = []byte("EqIN6cImvf7fUZJXK4YL")

type Claims struct {
	UserID       string
	UserRole     string
	UserNickname string
	jwt.StandardClaims
}

func GenerateToken(user *users.User) (string, error) {
	claims := &Claims{
		UserID:       user.ID.String(),
		UserRole:     user.Role,
		UserNickname: user.Nickname,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("SignedString: %s", err)
	}

	return tokenString, nil
}

func JwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("authorization")
		if tokenString == "" {
			return echo.ErrUnauthorized
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return secretKey, nil
		})

		UserRole := token.Claims.(jwt.MapClaims)["UserRole"]

		if err != nil || !token.Valid {
			return echo.ErrUnauthorized
		}
		if UserRole != "admin" {
			return echo.ErrForbidden
		}

		return next(c)
	}
}

func GetUserNicknameFromToken(t string) (string, error) {
	parts := strings.Split(t, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("Invalid token format")
	}
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(t, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return "", fmt.Errorf("parse error: %s", err)
	}
	userNickname, ok := claims["UserNickname"].(string)
	if !ok {
		return "", fmt.Errorf("nickname not found in claims")
	}
	return userNickname, nil
}
