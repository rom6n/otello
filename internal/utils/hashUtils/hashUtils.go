package hashutils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	saltLength = 12

	hashLength  = 32
	hashTime    = 2
	hashMemory  = 64 * 1024
	hashThreads = 4
)

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %v", err)
	}

	return salt, nil
}

func HashPassword(password string, salt []byte) string {
	hash := argon2.IDKey([]byte(password), salt, hashTime, hashMemory, hashThreads, hashLength)

	hashStr := base64.RawStdEncoding.EncodeToString(hash)
	saltStr := base64.RawStdEncoding.EncodeToString(salt)

	return fmt.Sprintf("%s.%s", saltStr, hashStr)
}

func VerifyPassword(password, hashedPassword string) error {
	data := strings.Split(hashedPassword, ".")
	if len(data) != 2 {
		return fmt.Errorf("incorrect hash string")
	}

	salt, decodeErr := base64.RawStdEncoding.DecodeString(data[0])
	if decodeErr != nil {
		return fmt.Errorf("failed to decode salt string: %v", decodeErr)
	}

	newPassword := HashPassword(password, salt)

	if newPassword != hashedPassword {
		return fmt.Errorf("wrong password")
	}

	return nil
}
