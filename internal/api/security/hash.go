package security

import "golang.org/x/crypto/bcrypt"

type HashModule interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) bool
}

type hashModule struct {
}

func NewHashModule() HashModule {
	return hashModule{}
}

func (h hashModule) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (h hashModule) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
