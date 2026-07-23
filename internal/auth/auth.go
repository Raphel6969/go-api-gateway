package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid authentication token")
	ErrExpiredToken = errors.New("token has expired")
)

type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type Claims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"`
}

type Authenticator struct {
	jwtSecret []byte
	apiKeys   map[string]string
}

func NewAuthenticator(jwtSecret string, validAPIKeys map[string]string) *Authenticator {
	return &Authenticator{
		jwtSecret: []byte(jwtSecret),
		apiKeys:   validAPIKeys,
	}
}

func (a *Authenticator) ValidateAPIKey(key string) (string, bool) {
	userID, exists := a.apiKeys[key]
	return userID, exists
}

func (a *Authenticator) ValidateJWT(tokenStr string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	signatureInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, a.jwtSecret)
	mac.Write([]byte(signatureInput))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if expectedSignature != parts[2] {
		return nil, ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > claims.Exp {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}
