package lib

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func CreateNanoid() (string, error) {
	id, err := gonanoid.New()
	if err != nil {
		return "", err
	}
	return id, nil
}
