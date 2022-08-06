package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"time"
)

var (
	ErrBadSigningMethod = errors.New("bad signing method")
	ErrBadToken         = errors.New("bad token")
	ErrInternalError    = errors.New("internal error")
	ErrEmptySigningKey  = errors.New("empty signing key")
)

type TokenManager struct {
	signingKey []byte
	tokenTTL   time.Duration
}

type customClaims struct {
	jwt.StandardClaims
	User interface{} `json:"user"`
}

func NewTokenManager(signingKey string, tokenTTL time.Duration) (*TokenManager, error) {
	if signingKey == "" {
		return nil, ErrEmptySigningKey
	}

	tm := &TokenManager{
		signingKey: []byte(signingKey),
		tokenTTL:   tokenTTL,
	}
	return tm, nil
}

func (m *TokenManager) Create(user interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(m.tokenTTL).Unix(),
		},
		User: user,
	})

	tokenString, err := token.SignedString(m.signingKey)
	if err != nil {
		return "", ErrInternalError
	}

	return tokenString, nil
}

func (m *TokenManager) Check(tokenString string) (interface{}, error) {
	signingKeyGetter := func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || method.Alg() != "HS256" {
			return nil, ErrBadSigningMethod
		}

		return m.signingKey, nil
	}

	claims := &customClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, signingKeyGetter)
	if err != nil || !token.Valid {
		return nil, ErrBadToken
	}

	return claims.User, nil
}
