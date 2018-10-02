package helpers

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var stubbedRunningDockerImages = []string{}

//
// Stub all calls to RunningDockerImages with a string slice.
// Useful for testing.
//
func StubRunningDockerImages(images []string) {
	stubbedRunningDockerImages = images
}

//
// Connect to the local docker socket and get a list of running docker images.
//
func RunningDockerImages() ([]string, error) {
	if len(stubbedRunningDockerImages) != 0 {
		return stubbedRunningDockerImages, nil
	}
	set := []string{}

	dockerAPI, err := client.NewEnvClient()
	if err != nil {
		return set, err
	}

	containers, err := dockerAPI.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return set, err
	}

	for _, container := range containers {
		set = append(set, container.Image)
	}
	return set, nil
}
