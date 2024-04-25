package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (TokenMaker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must have at least %v characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey}, nil
}

func (jwtMaker JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	return jwtToken.SignedString([]byte(jwtMaker.secretKey))
}

func (jwtMaker JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return []byte(jwtMaker.secretKey), nil
	}
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		// the jwt library hides this error in Inner, we have to check for it
		if ok {
			if errors.Is(verr.Inner, ErrExpiredToken) {
				return nil, ErrExpiredToken
			} else if errors.Is(verr.Inner, ErrInvalidToken) {
				return nil, ErrInvalidToken
			}
		} else {
			return nil, ErrInvalidToken
		}
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
