package facts

import (
	"github.com/coinbase/dexter/engine/helpers"
	"strings"
)

func init() {
	add(Fact{
		Name:             "running-docker-image-substring",
		Description:      "check if the host is running a docker container whos image contains the argument as a substring",
		MinimumArguments: 1,
		function:         runningDockerImageSubstring,
	})
	add(Fact{
		Name:             "running-docker-image",
		Description:      "check if the host is running a docker container based on the image provided as an argument",
		MinimumArguments: 1,
		function:         runningDockerImage,
	})
}

func runningDockerImage(args []string) (bool, error) {
	images, err := helpers.RunningDockerImages()
	if err != nil {
		return false, err
	}
	for _, arg := range args {
		for _, image := range images {
			if image == arg {
				return true, nil
			}
		}
	}
	return false, nil
}

func runningDockerImageSubstring(args []string) (bool, error) {
	images, err := helpers.RunningDockerImages()
	if err != nil {
		return false, err
	}
	for _, arg := range args {
		for _, image := range images {
			if strings.Contains(image, arg) {
				return true, nil
			}
		}
	}
	return false, nil
}
