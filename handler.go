package main

import (
	log "github.com/Sirupsen/logrus"

	docker "github.com/docker/engine-api/client"
)

type handler struct {
	*docker.Client
	Config *config
}

func NewHandler(version string) *handler {
	conf := loadConfig(version)
	c, err := docker.NewClient(conf.Docker.Endpoint, "", nil, nil)
	if err != nil {
		log.Fatal("failed to initialize Docker client")
	}
	return &handler{
		Client: c,
		Config: conf,
	}
}
