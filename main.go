package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

var version = "undefined"

func main() {
	var err error

	h := NewHandler(version)
	images, err := h.getImages()
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}

	for _, i := range images {
		err = i.pull()
		if err != nil {
			log.Errorf("Failed to pull image %s: %v", i, err)
		}

		err = i.kick()
		if err != nil {
			log.Errorf("Failed to kick containers for image %s: %v", i, err)
		}
	}
}
