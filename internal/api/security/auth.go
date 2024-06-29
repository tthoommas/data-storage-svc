package security

import (
	"data-storage-svc/internal/database"
)

func AuthenticateUser(Email *string, Password *string) bool {
	if Email == nil || len(*Email) == 0 || Password == nil || len(*Password) == 0 {
		return false
	}

	user, err := database.FindUserByEmail(Email)
	if err != nil {
		return false
	}
	return VerifyPassword(*Password, *user.PasswordHash)
}
