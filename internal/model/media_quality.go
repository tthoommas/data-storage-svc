package model

import (
	"math"
	"strings"
)

type MediaQuality uint

const (
	MICRO     MediaQuality = 150
	THUMBNAIL MediaQuality = 300
	MEDIUM    MediaQuality = 600
	HIGH      MediaQuality = 1200
	VERY_HIGH MediaQuality = 2400
	MAX       MediaQuality = math.MaxInt // Use maxInt  instead of maxUint to be able to cast to int when needed
)

func ParseMediaQuality(rawQuality string) MediaQuality {
	switch strings.ToLower(rawQuality) {
	case "micro":
		return MICRO
	case "thumbnail":
		return THUMBNAIL
	case "medium":
		return MEDIUM
	case "high":
		return HIGH
	case "very_high":
		return VERY_HIGH
	case "max":
		return MAX
	default:
		return MEDIUM
	}
}

func (m MediaQuality) AsUint() uint {
	return uint(m)
}

func (m MediaQuality) AsInt() int {
	return int(m)
}
