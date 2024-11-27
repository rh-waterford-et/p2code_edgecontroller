package utils

import (
	"errors"

	"github.com/google/go-containerregistry/pkg/crane"
)

const key = "bootc-agent-enabled"

func ValidateImage(imageRef string) error {
	image, err := crane.Pull(imageRef)

	if err != nil {
		return err
	} else {
		cf, err := image.ConfigFile()

		if err != nil {
			return err
		}

		imageLabels := cf.Config.Labels

		value, found := imageLabels[key]

		if !found || value != "true" {
			return errors.New("Invalid image - agent is missing")
		}
	}

	return nil
}
