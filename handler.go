package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	docker "github.com/docker/engine-api/client"
)

// Upkick is an upkick handler
type Upkick struct {
	*docker.Client
	Config *config
}

func newUpkick(version string) (*Upkick, error) {
	u := &Upkick{}
	err := u.setup(version)
	return u, err
}

func (u *Upkick) setup(version string) (err error) {
	u.Config = loadConfig(version)

	err = u.setupLoglevel()
	if err != nil {
		return errors.Wrap(err, "failed to setup log level")
	}

	err = u.setupDocker()
	if err != nil {
		return errors.Wrap(err, "failed to setup Docker")
	}

	return
}

func (u *Upkick) setupLoglevel() (err error) {
	switch u.Config.Loglevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		errMsg := fmt.Sprintf("Wrong log level '%v'", u.Config.Loglevel)
		err = errors.New(errMsg)
	}

	if u.Config.JSON {
		log.SetFormatter(&log.JSONFormatter{})
	}

	return
}

func (u *Upkick) setupDocker() (err error) {
	u.Client, err = docker.NewClient(u.Config.Docker.Endpoint, "", nil, nil)
	return
}
