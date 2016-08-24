package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

var client = &http.Client{Timeout: 10 * time.Second}

func loadMetrics(config rabbitExporterConfig, endpoint string) (*json.Decoder, error) {
	req, err := http.NewRequest("GET", config.RabbitURL+"/api/"+endpoint, nil)
	req.SetBasicAuth(config.RabbitUsername, config.RabbitPassword)

	resp, err := client.Do(req)

	if err != nil || resp == nil || resp.StatusCode != 200 {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		log.WithFields(log.Fields{"error": err, "host": config.RabbitURL, "statusCode": status}).Error("Error while retrieving data from rabbitHost")
		return nil, errors.New("Error while retrieving data from rabbitHost")
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return json.NewDecoder(bytes.NewBuffer(body)), nil
}

func getQueueInfo(config rabbitExporterConfig) ([]QueueInfo, error) {
	var q []QueueInfo

	d, err := loadMetrics(config, "queues")
	if err != nil {
		return q, err
	}

	q = MakeQueueInfo(d)

	return q, nil
}

func getOverviewMap(config rabbitExporterConfig) (MetricMap, error) {
	var overview MetricMap

	d, err := loadMetrics(config, "overview")
	if err != nil {
		return overview, err
	}

	return MakeMap(d), nil
}
