package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/upkick/handler"
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

	images, err := kicker.GetImages()
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}

	for _, i := range images {
		err = kicker.Pull(i)
		if err != nil {
			log.Errorf("Failed to pull image %s: %v", i, err)
		}

		err = kicker.Kick(i)
		if err != nil {
			log.Errorf("Failed to kick containers for image %s: %v", i, err)
		}
	}

	/* Metrics per node:
	 *
	 * - States (gauges):
	 *   - NOUP:     up-to-date containers
	 *   - UP-OK:    container successfully updated
	 *   - UP-NOK:   container failed to update
	 *   - OUT-WARN: container is out-of-date but not updated
	 *
	 * - Image timestamp per hash (counter)
	 */
	kicker.Metrics.Push()
}
