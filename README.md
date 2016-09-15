Upkick
======

[![Docker Pulls](https://img.shields.io/docker/pulls/camptocamp/upkick.svg)](https://hub.docker.com/r/camptocamp/upkick/)
[![Build Status](https://img.shields.io/travis/camptocamp/upkick/master.svg)](https://travis-ci.org/camptocamp/upkick)
[![Coverage Status](https://img.shields.io/coveralls/camptocamp/upkick.svg)](https://coveralls.io/r/camptocamp/upkick?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/camptocamp/upkick)](https://goreportcard.com/report/github.com/camptocamp/upkick)
[![By Camptocamp](https://img.shields.io/badge/by-camptocamp-fb7047.svg)](http://www.camptocamp.com)


Unattended upgrades for Docker containers, the hard way.


## Installing

```shell
$ go get github.com/camptocamp/upkick
```

## Usage

```shell
Usage:
  upkick [OPTIONS]

Application Options:
  -V, --version          Display version.
  -l, --loglevel=        Set loglevel ('debug', 'info', 'warn', 'error', 'fatal', 'panic'). (default: info) [$UPKICK_LOG_LEVEL]
  -m, --manpage          Output manpage.
  -j, --json             Log as JSON (to stderr). [$UPKICK_JSON_OUTPUT]
  -w, --warn-only        Only warn, do not kick out-of-date containers. [$UPKICK_WARN_ONLY]

Docker Options:
  -e, --docker-endpoint= The Docker endpoint. (default: unix:///var/run/docker.sock) [$DOCKER_ENDPOINT]

Help Options:
  -h, --help             Show this help message
```

## Using Docker

```shell
$ docker run -v /var/run/docker.sock:/var/run/docker.sock:ro  --rm -ti \
     camptocamp/upkick
```

## Per container opt-out

You can set containers to only warn if they are outdated by placing an `io.upkick.warn_only=true` label on them.
