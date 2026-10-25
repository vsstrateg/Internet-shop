package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("недействительный токен")
	ErrExpiredToken = errors.New("срок действия токена истек")
)

func (a *AuthService) GenerateJWT(email string) (string, error) {

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "MyAuthServer",
		"sub": email,
		"exp": time.Now().Add(20 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	})

	s, err := claims.SignedString(a.jwtSecret)
	if err != nil {
		return "", err
	}
	return s, nil
}

func (a *AuthService) ParseJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return a.jwtSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}
		return "", ErrInvalidToken
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		email, ok := claims["sub"].(string)
		if !ok {
			return "", ErrInvalidToken
		}
		return email, nil
	}
	return "", ErrInvalidToken
}
