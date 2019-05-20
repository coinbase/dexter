package tasks

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/coinbase/dexter/engine/helpers/docker"

	log "github.com/sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
)

func init() {
	add(Task{
		Name:                 "docker-filesystem-diff",
		Description:          "collect artifacts and a report on all changes to docker container filesystems",
		ConsensusRequirement: 1,
		actionFunction:       exportContainerFilesystemDiffReport,
	})
}

type containerDiffType int

const (
	containerDiffModified = iota
	containerDiffAdded
	containerDiffRemoved
)

var diffName = map[containerDiffType]string{
	0: "Modified",
	1: "Added",
	2: "Removed",
}

type containerFilesystemChange struct {
	ChangeType string
	DiffType   containerDiffType `json:"-"`
	Path       string
	PathStat   types.ContainerPathStat
}

type containerChangeSet struct {
	Container types.Container
	Changes   []containerFilesystemChange
}

func exportContainerFilesystemDiffReport(_ []string, writer *ArtifactWriter) {
	allContainers, err := docker.API().ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		errstr := "unable to list containers for task"
		log.WithFields(log.Fields{
			"at":    "actions.ExportContainerFilesystemDiffReport",
			"error": err.Error(),
		}).Error(errstr)
		writer.Error(errstr + ": " + err.Error())
		return
	}
	for _, container := range allContainers {
		report, err := containerDiff(writer, docker.API(), container)
		if err != nil {
			errstr := "error creating container diff"
			log.WithFields(log.Fields{
				"at":    "actions.ExportContainerFilesystemDiffReport",
				"error": err.Error(),
			}).Error(errstr)
			writer.Error(errstr + ": " + err.Error())
			continue
		}
		zipChanges(writer, report)
	}
}

func containsRemovedOrAdded(report containerChangeSet) bool {
	for _, change := range report.Changes {
		switch change.DiffType {
		case containerDiffRemoved:
			return true
		case containerDiffModified:
			return true
		}
	}
	return false
}

func zipChanges(writer *ArtifactWriter, report containerChangeSet) {
	// Write a high-level manifest of the changes
	writeContainerManifest(writer, report)

	// Start a container from the original image for extracting removed files
	tmpContainer := ""
	if containsRemovedOrAdded(report) {
		var err error
		tmpContainer, err = startOriginalContainer(writer, docker.API(), report.Container.Image)
		if err != nil {
			errstr := "error starting original container for image"
			log.WithFields(log.Fields{
				"at":    "actions.zipChanges",
				"error": err.Error(),
				"image": report.Container.Image,
			}).Error(errstr)
			writer.Error(errstr + " (" + report.Container.Image + ") : " + err.Error())
		}
	}

	// Extract all added/removed/modified files from container
	for _, change := range report.Changes {
		switch change.DiffType {
		case containerDiffAdded:
			writeAddedFile(writer, report.Container.ID, change)
		case containerDiffRemoved:
			writeRemovedFile(writer, report.Container.ID, tmpContainer, change)
		case containerDiffModified:
			writeModifiedFile(writer, report.Container.ID, tmpContainer, change)
		}
	}

	// Cleanup unchanged container made from original image
	if tmpContainer != "" {
		docker.API().ContainerKill(context.TODO(), tmpContainer, "SIGKILL")
	}
}

func writeAddedFile(writer *ArtifactWriter, container string, change containerFilesystemChange) {
	data, err := extractFile(writer, docker.API(), change.Path, container)
	if err != nil {
		errstr := "error extracting file from container"
		log.WithFields(log.Fields{
			"at":        "actions.writeAddedFile",
			"error":     err.Error(),
			"path":      change.Path,
			"container": container,
		}).Error(errstr)
		writer.Error(errstr + " (" + container + " " + change.Path + ") :" + err.Error())
		return
	}
	writer.Write(container+"/added"+change.Path, data)
}

func writeRemovedFile(writer *ArtifactWriter, container, tmpContainer string, change containerFilesystemChange) {
	data, err := extractFile(writer, docker.API(), change.Path, tmpContainer)
	if err != nil {
		errstr := "error extracting file from container"
		log.WithFields(log.Fields{
			"at":        "actions.writeRemovedFile",
			"error":     err.Error(),
			"path":      change.Path,
			"container": container,
		}).Error(errstr)
		writer.Error(errstr + " (" + container + " " + change.Path + ") :" + err.Error())
		return
	}
	writer.Write(container+"/removed"+change.Path, data)
}

func writeModifiedFile(writer *ArtifactWriter, container, tmpContainer string, change containerFilesystemChange) {
	// Extract modified file
	mdata, err := extractFile(writer, docker.API(), change.Path, container)
	if err != nil {
		errstr := "error extracting modified file from container"
		log.WithFields(log.Fields{
			"at":        "actions.writeModifiedFile",
			"error":     err.Error(),
			"path":      change.Path,
			"container": container,
		}).Error(errstr)
		writer.Error(errstr + " (" + container + " " + change.Path + ") :" + err.Error())
	} else {
		writer.Write(container+"/modified"+change.Path, mdata)
	}
	// Extract original file
	odata, err := extractFile(writer, docker.API(), change.Path, tmpContainer)
	if err != nil {
		errstr := "error extracting original file from container"
		log.WithFields(log.Fields{
			"at":        "actions.writeModifiedFile",
			"error":     err.Error(),
			"path":      change.Path,
			"container": container,
		}).Error(errstr)
		writer.Error(errstr + " (" + container + " " + change.Path + ") :" + err.Error())
	} else {
		writer.Write(container+"/modified"+change.Path+".original", odata)
	}
}

func extractFile(writer *ArtifactWriter, cli *client.Client, path, container string) ([]byte, error) {
	readCloser, _, err := cli.CopyFromContainer(context.TODO(), container, path)
	if err != nil {
		errstr := "unable to pull file out of container"
		log.WithFields(log.Fields{
			"at":        "actions.extractFile",
			"error":     err.Error(),
			"container": container,
			"path":      path,
		}).Error(errstr)
		writer.Error(errstr + " (" + container + " " + path + ") :" + err.Error())
		return []byte{}, err
	}
	// The tarData contains a tar archive from the docker API.  This must be unpacked to get to the actual
	// file contained in the tar archive. The tar is scanned until a file type is found (not a directory)
	// and then the file contents are assigned to the variable that is written into the zip.
	tarReader := tar.NewReader(readCloser)
	buffer := new(bytes.Buffer)
	lastRun := false
	for !lastRun {
		header, err := tarReader.Next()
		if err == io.EOF {
			lastRun = true
		}
		// Anything except EOF should cause a bail on this loop
		if (err != nil && err != io.EOF) || header == nil {
			errstr := "error getting next header from tar reader"
			log.WithFields(log.Fields{
				"at":    "actions.extractFile",
				"error": err.Error(),
			}).Error(errstr)
			writer.Error(errstr + ": " + err.Error())
			return []byte{}, err
		}
		if header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeRegA {
			// This is the requested file
			_, err := io.Copy(buffer, tarReader)
			if err != nil {
				errstr := "error copying file from container"
				log.WithFields(log.Fields{
					"at":    "tasks.extractFile",
					"error": err.Error(),
				}).Error(errstr)
				writer.Error(errstr + ": " + err.Error())
			}
			break
		}
	}
	readCloser.Close()
	return buffer.Bytes(), nil
}

func startOriginalContainer(writer *ArtifactWriter, cli *client.Client, image string) (string, error) {
	response, err := cli.ContainerCreate(
		context.TODO(),
		&container.Config{
			Entrypoint:      strslice.StrSlice{"/bin/sleep", "900"},
			Healthcheck:     &container.HealthConfig{Test: []string{"NONE"}},
			Image:           image,
			NetworkDisabled: true,
		},
		&container.HostConfig{
			AutoRemove: true,
		},
		&network.NetworkingConfig{},
		"",
	)
	if err != nil {
		errstr := "error starting original container for image"
		log.WithFields(log.Fields{
			"at":    "actions.startOriginalContainer",
			"error": err.Error(),
		}).Error(errstr)
		writer.Error(errstr + ": " + err.Error())
		return "", err
	}
	for _, message := range response.Warnings {
		warnstr := "warning starting original container"
		log.WithFields(log.Fields{
			"at":      "actions.startOriginalContainer",
			"context": warnstr,
		}).Warn(message)
		writer.Error(warnstr + ": " + message)
	}
	return response.ID, nil
}

func writeContainerManifest(writer *ArtifactWriter, report containerChangeSet) {
	manifestFile := report.Container.ID + "/manifest.json"
	manifestData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		errstr := "error creating container diff manifest"
		log.WithFields(log.Fields{
			"at":    "actions.writeContainerManifest",
			"error": err.Error(),
		}).Error(errstr)
		writer.Error(errstr + ": " + err.Error())
		return
	}
	writer.Write(manifestFile, manifestData)
}

func containerDiff(writer *ArtifactWriter, cli *client.Client, container types.Container) (changeSet containerChangeSet, err error) {
	responses, err := cli.ContainerDiff(context.Background(), container.ID)
	if err != nil {
		errstr := "error calling cli.containerDiff"
		log.WithFields(log.Fields{
			"at":        "actions.containerDiff",
			"error":     err.Error(),
			"container": container.ID,
		}).Error(errstr)
		writer.Error(errstr + "in continer " + container.ID + ": " + err.Error())
		return
	}

	changeSet.Container = container

	for _, response := range responses {
		pathStat, serr := cli.ContainerStatPath(context.Background(), container.ID, response.Path)

		// Not found errors are expected on removed files
		if response.Kind != containerDiffRemoved && serr != nil {
			err = serr
			errstr := "error calling cli.ContainerStatPath"
			log.WithFields(log.Fields{
				"at":        "actions.containerDiff",
				"error":     err.Error(),
				"container": container.ID,
				"path":      response.Path,
			}).Error(errstr)
			writer.Error(errstr + "in continer " + container.ID + ", path " + response.Path + ": " + err.Error())
			return
		}

		// Exclude directories as every directory in the path of a changed file will appear changed.
		if pathStat.Mode.IsDir() {
			continue
		}

		changeSet.Changes = append(changeSet.Changes, containerFilesystemChange{
			ChangeType: diffName[containerDiffType(response.Kind)],
			DiffType:   containerDiffType(response.Kind),
			Path:       response.Path,
			PathStat:   pathStat,
		})
	}
	return
}
