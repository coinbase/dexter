package docker

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/docker/docker/client"
)

var cachedAPI *client.Client

//
// Connect to the local docker socket, returning the cached client
// if a connection has already been made.
//
func API() *client.Client {
	if cachedAPI != nil {
		return cachedAPI
	}
	var dockerAPI *client.Client
	var err error
	var connected = false
	for !connected {
		dockerAPI, err = client.NewEnvClient()
		if err != nil {
			log.WithFields(log.Fields{
				"at":    "engine.Start",
				"error": err.Error(),
			}).Error("failed to connect to docker daemon, retry in 10s")
			time.Sleep(10 * time.Second)
		} else {
			connected = true
			cachedAPI = dockerAPI
		}
	}
	return dockerAPI
}
