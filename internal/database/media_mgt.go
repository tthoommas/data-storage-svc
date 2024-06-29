package database

import (
	"context"
	"data-storage-svc/internal/model"
)

const (
	MEDIA_COLLECTION = "medias"
)

func CreateMedia(Media *model.Media) error {
	_, err := Mongo().Collection(MEDIA_COLLECTION).InsertOne(context.Background(), Media)
	return err
}
