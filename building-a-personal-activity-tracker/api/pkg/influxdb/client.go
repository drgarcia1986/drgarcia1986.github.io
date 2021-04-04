package influxdb

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type realClient struct {
	endpoint string
	database string
}

func (r *realClient) Write(class, title string) error {
	resp, err := http.Post(
		fmt.Sprintf("%s/write?db=%s", r.endpoint, r.database),
		"",
		buildPayload(class, title),
	)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.New(
				fmt.Sprintf("Error reading influxdb response: %+v", err),
			)
		}
		return errors.New(
			fmt.Sprintf(
				"Invalid influxdb response: %d - %s",
				resp.StatusCode,
				string(body),
			),
		)
	}
	return nil
}

func NewClient(endpoint, database string) Client {
	return &realClient{
		endpoint: endpoint,
		database: database,
	}
}
