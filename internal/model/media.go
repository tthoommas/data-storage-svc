package model

import "time"

type Media struct {
	OriginalFileName *string    `json:"originalFileName"`
	StorageFileName  *string    `json:"storageFileName"`
	UploadedBy       *string    `json:"uploadedBy"`
	UploadTime       *time.Time `json:"uploadTime"`
}
