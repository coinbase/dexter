package tasks

import (
	"github.com/coinbase/dexter/engine/helpers"

	log "github.com/Sirupsen/logrus"
	"github.com/kolide/osquery-go"

	"encoding/json"
	"fmt"
	"time"
)

func init() {
	add(Task{
		Name:             "osquery-collect",
		Description:      "collect all data from all tables in osquery",
		MinimumArguments: 0,
		actionFunction:   collectOSQuery,
	})
}

func collectOSQuery(_ []string, writer *ArtifactWriter) {
	socket := helpers.OSQuerySocket()
	client, err := osquery.NewClient(socket, 60*time.Second)
	if err != nil {
		errstr := "error creating osquery client"
		log.WithFields(log.Fields{
			"at":    "tasks.collectOSQuery",
			"error": err.Error(),
		}).Error(errstr)
		writer.Error(errstr + ": " + err.Error())
		return
	}
	defer client.Close()

	tableNames := getTables(client, writer)
	for _, table := range tableNames {
		query := fmt.Sprintf("SELECT * FROM %s;", table)
		resp, err := client.Query(query)
		if err != nil {
			errstr := "error running query against osquery"
			log.WithFields(log.Fields{
				"at":    "tasks.collectOSQuery",
				"error": err.Error(),
				"query": query,
			}).Error(errstr)
			writer.Error(errstr + " (" + query + ") :" + err.Error())
			continue
		}
		if resp.Status.Code != 0 {
			errstr := "query returned non-zero response"
			log.WithFields(log.Fields{
				"at":       "tasks.collectOSQuery",
				"response": resp.Status.Message,
				"query":    query,
			}).Error(errstr)
			writer.Error(errstr + " (" + query + ") :" + err.Error())
			continue
		}
		data, err := json.MarshalIndent(resp.Response, "", "    ")
		if err != nil {
			errstr := "failed to json marshal osquery result"
			log.WithFields(log.Fields{
				"at":    "tasks.collectOSQuery",
				"error": err.Error(),
			}).Error(errstr)
			writer.Error(errstr + ": " + err.Error())
			continue
		}
		writer.Write(
			table+"/results.json",
			data,
		)
	}
}

func getTables(client *osquery.ExtensionManagerClient, writer *ArtifactWriter) []string {
	set := []string{}
	query := `SELECT name FROM sqlite_temp_master WHERE type="table";`
	resp, err := client.Query(query)
	if err != nil {
		errstr := "error running query against osquery"
		log.WithFields(log.Fields{
			"at":    "tasks.getTables",
			"error": err.Error(),
			"query": query,
		}).Error(errstr)
		writer.Error(errstr + " (" + query + ") :" + err.Error())
		return set
	}
	if resp.Status.Code != 0 {
		errstr := "query returned non-zero response"
		log.WithFields(log.Fields{
			"at":       "tasks.getTables",
			"response": resp.Status.Message,
			"query":    query,
		}).Error(errstr)
		return set
	}
	for _, tabledef := range resp.Response {
		if tableName, ok := tabledef["name"]; ok {
			set = append(set, tableName)
		} else {
			errstr := "response from table query did not contain name field"
			log.WithFields(log.Fields{
				"at": "tasks.getTables",
			}).Error(errstr)
			writer.Error(errstr + ": " + err.Error())
		}
	}
	return set
}
