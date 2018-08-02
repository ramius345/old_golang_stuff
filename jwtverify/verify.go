package jwtverify

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
)

type SigningKey struct {
	Signing_key string
}

type ISiginingKey interface {
	Verify(token string) (bool, *jwt.MapClaims)
}

func (k SigningKey) Verify(tokenString string) (bool, *jwt.MapClaims) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(k.Signing_key), nil
	})

	if err != nil {
		return false, nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true, &claims
	}

	return false, nil
}
