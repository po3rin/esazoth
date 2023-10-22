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
	"github.com/po3rin/esazoth/entity"
)

type ESConfig struct {
	Username string `default:""`
	Password string `default:""`
	Endpoint string `default:"http://localhost:9200"`
}

type ReindexWaitForCompletionResponse struct {
	Task string `json:"task"`
}

type ES struct {
	client   *http.Client
	endpoint string
	username string
	password string
}

type ESTaskResponse struct {
	Completed bool               `json:"completed"`
	Task      ESTaskResponseTask `json:"task"`
}

type ESTaskResponseTask struct {
	Node               string               `json:"node"`
	ID                 int                  `json:"id"`
	Status             ESTaskResponseStatus `json:"status"`
	StartTimeInMillis  uint64               `json:"start_time_in_millis"`
	RunningTimeInNanos uint64
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

func (e *ES) Task(ctx context.Context, id string) (entity.Task, error) {
	endpoint, err := url.JoinPath(e.endpoint, "_tasks", id)
	if err != nil {
		return entity.Task{}, err
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return entity.Task{}, err
	}

	req.SetBasicAuth(e.username, e.password)
	req.Header.Set("Content-Type", "application/json")

	if ctx != nil {
		req.WithContext(ctx)
	}

	res, err := e.client.Do(req)
	if err != nil {
		return entity.Task{}, err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return entity.Task{}, err
	}

	var r ESTaskResponse
	if err := json.Unmarshal(b, &r); err != nil {
		return entity.Task{}, err
	}

	return entity.Task{ID: fmt.Sprintf("%v:%v", r.Task.Node, r.Task.ID), Completed: r.Completed, StartTimeInMillis: r.Task.StartTimeInMillis}, nil
}

func (e *ES) ReindexWaitForCompletion(ctx context.Context, body io.Reader) (string, error) {
	endpoint, err := url.JoinPath(e.endpoint, "_reindex")
	endpoint = fmt.Sprintf("%s?wait_for_completion=false", endpoint)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(e.username, e.password)
	req.Header.Set("Content-Type", "application/json")

	if ctx != nil {
		req.WithContext(ctx)
	}

	res, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var r ReindexWaitForCompletionResponse
	if err := json.Unmarshal(b, &r); err != nil {
		return "", err
	}

	return r.Task, nil
}
