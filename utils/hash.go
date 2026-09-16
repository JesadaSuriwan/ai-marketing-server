package utils

import (
	"crypto/rand"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPassword(password, hashPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}

// tempPasswordAlphabet excludes visually-ambiguous characters (0/O, 1/l/I)
// since this is meant to be read off a screen and typed by a human.
const tempPasswordAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789"

// GenerateTempPassword produces a random 12-character human-typeable
// password for accounts created on someone else's behalf (e.g. a team
// member added by an Admin/Team Lead), who must change it on first login.
func GenerateTempPassword() (string, error) {
	const length = 12
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, length)
	for i, v := range b {
		out[i] = tempPasswordAlphabet[int(v)%len(tempPasswordAlphabet)]
	}
	return string(out), nil
}
