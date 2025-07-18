package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func HashMedia(filename string) (*string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	hasher := sha256.New()
	buffer := make([]byte, hasher.BlockSize())
	var readBytes int
	for err != io.EOF {
		if readBytes, err = file.Read(buffer); err != nil && err != io.EOF {
			return nil, err
		}
		if _, hashErr := hasher.Write(buffer[0:readBytes]); hashErr != nil {
			return nil, hashErr
		}
	}
	hashResult := hex.EncodeToString(hasher.Sum(nil))
	return &hashResult, nil
}
