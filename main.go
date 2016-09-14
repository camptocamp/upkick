package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

var version = "undefined"
var upkick *Upkick

func init() {
	var err error
	upkick, err = newUpkick(version)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	var err error

	images, err := upkick.getImages()
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
