package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// EncryptPassword/DecryptPassword back the admin-facing "view/edit a Customer
// role member's password" feature — a deliberate, separate reversible store
// alongside the normal bcrypt hash used for actual login. AES-256-GCM: a
// random 12-byte nonce is generated per call and stored alongside the
// ciphertext (nonce || ciphertext, base64-encoded as one string) since GCM
// needs it for decryption and a nonce isn't secret.
//
// key must be 32 raw bytes, given as a base64 string (MEMBER_PASSWORD_ENCRYPTION_KEY).
func EncryptPassword(plaintext, base64Key string) (string, error) {
	block, err := newAESCipher(base64Key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func DecryptPassword(encoded, base64Key string) (string, error) {
	block, err := newAESCipher(base64Key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	sealed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(sealed) < nonceSize {
		return "", errors.New("encrypted password is malformed")
	}
	nonce, ciphertext := sealed[:nonceSize], sealed[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func newAESCipher(base64Key string) (cipher.Block, error) {
	if base64Key == "" {
		return nil, errors.New("MEMBER_PASSWORD_ENCRYPTION_KEY is not configured")
	}
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, errors.New("MEMBER_PASSWORD_ENCRYPTION_KEY is not valid base64")
	}
	if len(key) != 32 {
		return nil, errors.New("MEMBER_PASSWORD_ENCRYPTION_KEY must decode to exactly 32 bytes (AES-256)")
	}
	return aes.NewCipher(key)
}
