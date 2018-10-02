package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/coinbase/dexter/engine/helpers/docker"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
)

func (investigation *Investigation) cleanup() {
	investigation.removeReportArtifacts()
	if investigation.KillContainers {
		killContainers()
	}
	log.WithFields(log.Fields{
		"at":            "engine.cleanup",
		"investigation": investigation.ID,
	}).Info("investigation complete")
	if investigation.KillHost {
		shutdownHost()
	}
}

func (investigation *Investigation) removeReportArtifacts() {
	err := os.RemoveAll(filepath.FromSlash(investigation.ReportDirectory()))
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.removeReportArtifacts",
			"error": err.Error(),
		}).Error("error removing report directory")
	}
	err = nil
	err = os.Remove(investigation.ReportZip())
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.removeReportArtifacts",
			"error": err.Error(),
		}).Error("error removing report zip")
	}
	err = nil
	err = os.Remove(investigation.ReportZip() + ".enc")
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.removeReportArtifacts",
			"error": err.Error(),
		}).Error("error removing encrypted report zip")
	}
}

func killContainers() {
	containers, err := docker.API().ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Error("unable to list containers to kill")
		return
	}
	for _, container := range containers {
		// Don't stop dexter containers
		if strings.Contains(container.Image, "dexter") {
			continue
		}
		log.WithFields(log.Fields{
			"at":           "killContainers",
			"container_id": container.ID,
		}).Info("killing container")
		docker.API().ContainerKill(context.TODO(), container.ID, "SIGKILL")
	}
}

func shutdownHost() {
	log.Info("dexter shutting down host")
	attrs := syscall.ProcAttr{
		Dir:   "",
		Env:   []string{},
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		Sys:   nil}
	var ws syscall.WaitStatus
	pid, err := syscall.ForkExec(
		"/sbin/shutdown",
		[]string{"shutdown", "-h", "now"},
		&attrs)
	if err != nil {
		log.Error(err)
		return
	}
	_, err = syscall.Wait4(pid, &ws, 0, nil)
	if err != nil {
		log.Error(err)
		return
	}
	os.Exit(0)
}
