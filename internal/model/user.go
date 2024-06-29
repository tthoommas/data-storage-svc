package model

import "time"

type User struct {
	Email              *string
	JoinDate           *time.Time
	PasswordHash       *string
	GrantedPermissions []int
}
