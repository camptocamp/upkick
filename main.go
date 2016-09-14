package main

import (
	log "github.com/Sirupsen/logrus"
)

func main() {
	var err error

	for _, i := range getImages() {
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
