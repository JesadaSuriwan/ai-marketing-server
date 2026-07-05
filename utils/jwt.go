package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

func GenerateToken(email string, userId int, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":  email,
		"userId": userId,
		"role":   role,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})

	secretKey := viper.GetString("env.secretkey")

	return token.SignedString([]byte(secretKey))
}

func VerifyToken(token string) (int, string, error) {
	secretKey := viper.GetString("env.secretkey")
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)

		if !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(secretKey), nil
	})

	if err != nil {
		return 0, "", errors.New("could not parse token")
	}

	tokenIsValid := parsedToken.Valid

	if !tokenIsValid {
		return 0, "", errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)

	if !ok {
		return 0, "", errors.New("invalid token claims")
	}

	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		return 0, "", errors.New("token expired")
	}

	userIdFloat := claims["userId"].(float64)
	userId := int(userIdFloat)

	role := claims["role"].(string)

	return userId, role, nil
}
