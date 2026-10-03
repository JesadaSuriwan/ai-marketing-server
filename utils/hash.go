package utils

import (
	"crypto/rand"
	"math/big"

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

// Each class excludes visually-ambiguous characters (0/O, 1/l/I) since this
// is meant to be read off a screen and typed by a human.
const (
	tempPasswordUpper   = "ABCDEFGHJKMNPQRSTUVWXYZ"
	tempPasswordLower   = "abcdefghjkmnpqrstuvwxyz"
	tempPasswordDigits  = "23456789"
	tempPasswordSymbols = "!@#$%^&*-_=+"
)

var tempPasswordAlphabet = tempPasswordUpper + tempPasswordLower + tempPasswordDigits + tempPasswordSymbols

// randomChar picks a uniformly random character from alphabet via
// crypto/rand.Int — unlike `randomByte % len(alphabet)`, this has no modulo
// bias toward the start of the alphabet.
func randomChar(alphabet string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
	if err != nil {
		return 0, err
	}
	return alphabet[n.Int64()], nil
}

// GenerateTempPassword produces a random 12-character human-typeable
// password for accounts created on someone else's behalf — a team member
// added by an Admin/Team Lead, or a Customer role member's password.
// Guarantees at least one uppercase letter, one lowercase letter, one digit,
// and one symbol (placed at random positions, not always the first four),
// satisfying typical password-complexity requirements outright.
func GenerateTempPassword() (string, error) {
	const length = 12
	categories := []string{tempPasswordUpper, tempPasswordLower, tempPasswordDigits, tempPasswordSymbols}

	out := make([]byte, length)
	for i, cat := range categories {
		c, err := randomChar(cat)
		if err != nil {
			return "", err
		}
		out[i] = c
	}
	for i := len(categories); i < length; i++ {
		c, err := randomChar(tempPasswordAlphabet)
		if err != nil {
			return "", err
		}
		out[i] = c
	}

	// Fisher-Yates shuffle so the four guaranteed characters aren't always
	// sitting in the first four positions.
	for i := length - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		j := jBig.Int64()
		out[i], out[j] = out[j], out[i]
	}

	return string(out), nil
}
