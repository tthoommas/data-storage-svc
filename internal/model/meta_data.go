package model

import "time"

type MetaData struct {
	Created     time.Time `json:"created,omitempty"`
	CameraModel string    `json:"cameraModel,omitempty"`
	Location    *Location `json:"location,omitempty"`
}

type Location struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Altitude  float32 `json:"altitude,omitempty"`
}
