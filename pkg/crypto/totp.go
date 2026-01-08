package crypto

import (
	"crypto/rand"
	"encoding/base32"
	"time"

	"github.com/pquerna/otp/totp"
)

type TOTPManager struct {
	Issuer string
}

func NewTOTPManager(issuer string) *TOTPManager {
	return &TOTPManager{
		Issuer: issuer,
	}
}

func (m *TOTPManager) GenerateSecret() (string, error) {
	secretSize := 20
	secret := make([]byte, secretSize)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(secret), nil
}

func (m *TOTPManager) GenerateCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}

func (m *TOTPManager) ValidateCode(secret, token string) bool {
	return totp.Validate(token, secret)
}

func (m *TOTPManager) ValidateCodeWithWindow(secret, token string, window int) bool {
	for i := -window; i <= window; i++ {
		t := time.Now().Add(time.Duration(i) * 30 * time.Second)
		code, _ := totp.GenerateCodeCustom(secret, t, totp.ValidateOpts{
			Period:    30,
			Skew:      0,
			Digits:    6,
			Algorithm: 0,
		})
		if code == token {
			return true
		}
	}
	return false
}
