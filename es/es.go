package es

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type ESConfig struct {
	Username string `default:""`
	Password string `default:""`
	Endpoint string `default:"http://localhost:9200"`
}

type ES struct {
	client   *http.Client
	endpoint string
	username string
	password string
}

type ESTaskResponse struct {
	Completed bool                    `json:"completed"`
	Task      ESTaskResponseTask      `json:"task"`
	Response  *ESTaskResponseResponse `json:"response"`
	Error     *ESTaskResponseError    `json:"error"`
}

type ESTaskResponseTask struct {
	Node               string               `json:"node"`
	ID                 int                  `json:"id"`
	Status             ESTaskResponseStatus `json:"status"`
	StartTimeInMillis  uint64               `json:"start_time_in_millis"`
	RunningTimeInNanos uint64               `json:"running_time_nanos"`
}

type ESTaskResponseResponse struct {
	TimeOut bool   `json:"timed_out"`
	Total   uint64 `json:"total"`
}

type ESTaskResponseError struct {
	Type   string `json:"type"`
	Reason int    `json:"reason"`
}

type ESTaskResponseStatus struct {
	Total   int `json:"total"`
	Updated int `json:"updated"`
	Created int `json:"created"`
	Deleted int `json:"deleted"`
	Batches int `json:"batches"`
}

func NewClient() (*ES, error) {
	var conf ESConfig
	err := envconfig.Process("ESAZOTH_ES", &conf)
	if err != nil {
		return nil, fmt.Errorf("DB environment is not correct: %w", err)
	}

	client := &http.Client{Timeout: time.Duration(30) * time.Second}
	return &ES{
		client:   client,
		endpoint: conf.Endpoint,
		username: conf.Username,
		password: conf.Password,
	}, nil
}

func (e *ES) Task(ctx context.Context, id string) (*ESTaskResponse, error) {
	endpoint, err := url.JoinPath(e.endpoint, "_tasks", id)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(e.username, e.password)
	req.Header.Set("Content-Type", "application/json")

	if ctx != nil {
		req.WithContext(ctx)
	}

	res, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var r ESTaskResponse
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}

	return &r, nil
}
