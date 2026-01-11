package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vachanmn123/vachancms/config"
)

type Claims struct {
	UserID          string `json:"user_id"`
	UserAccessToken string `json:"user_access_token"`
	jwt.RegisteredClaims
}

func encryptToken(token string) string {
	b64Token := base64.StdEncoding.EncodeToString([]byte(token))
	// AES encryption
	key := config.Cfg.EncryptionKey

	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(b64Token), nil))
}

func decryptToken(encryptedToken string) (string, error) {
	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", err
	}

	key := config.Cfg.EncryptionKey

	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedBytes) < nonceSize {
		return "", errors.New("invalid encrypted token")
	}

	nonce, ciphertext := encryptedBytes[:nonceSize], encryptedBytes[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	b64Token := string(plaintext)
	tokenBytes, err := base64.StdEncoding.DecodeString(b64Token)
	if err != nil {
		return "", err
	}

	return string(tokenBytes), nil
}

func GenerateJWT(userID string, userAccessToken string, cfg *config.Config) (string, error) {
	claims := Claims{
		UserID:          userID,
		UserAccessToken: encryptToken(userAccessToken),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

func ValidateJWT(tokenString string, cfg *config.Config) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {

		if decryptedToken, err := decryptToken(claims.UserAccessToken); err == nil {
			claims.UserAccessToken = decryptedToken
		}
		return claims, nil
	}
	return nil, err
}
