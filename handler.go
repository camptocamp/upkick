package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	docker "github.com/docker/engine-api/client"
)

type handler struct {
	*docker.Client
	Config *config
}

func newHandler(version string) (*handler, error) {
	h := &handler{}
	err := h.setup(version)
	return h, err
}

func (h *handler) setup(version string) (err error) {
	h.Config = loadConfig(version)

	err = h.setupLoglevel()
	if err != nil {
		return errors.Wrap(err, "failed to setup log level")
	}

	err = h.setupDocker()
	if err != nil {
		return errors.Wrap(err, "failed to setup Docker")
	}

	return
}

func (h *handler) setupLoglevel() (err error) {
	switch h.Config.Loglevel {
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
		errMsg := fmt.Sprintf("Wrong log level '%v'", h.Config.Loglevel)
		err = errors.New(errMsg)
	}

	if h.Config.JSON {
		log.SetFormatter(&log.JSONFormatter{})
	}

	return
}

func (h *handler) setupDocker() (err error) {
	h.Client, err = docker.NewClient(h.Config.Docker.Endpoint, "", nil, nil)
	return
}
