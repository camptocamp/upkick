package main

import (
	"os"

	"github.com/camptocamp/upkick/handler"
	log "github.com/sirupsen/logrus"
)

var version = "undefined"
var kicker *handler.Upkick

func init() {
	var err error
	kicker, err = handler.NewUpkick(version)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	var err error

	log.Infof("Upkick v%s starting", version)

	images, err := kicker.GetImages()
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}

	for _, i := range images {
		err = kicker.Pull(i)
		if err != nil {
			log.Errorf("Failed to pull image %s: %v", i, err)
			continue
		}

		err = kicker.Kick(i)
		if err != nil {
			log.Errorf("Failed to kick containers for image %s: %v", i, err)
		}
	}

	kicker.PushMetrics()
}
