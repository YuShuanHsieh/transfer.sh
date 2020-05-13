package apiauth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type APIConfig struct {
	Endpoint string
	Headers  map[string]string
}

type APIAuthenticator struct {
	config APIConfig
}

func New(cfg APIConfig) *APIAuthenticator {
	return &APIAuthenticator{
		config: cfg,
	}
}

func (a *APIAuthenticator) Authenticate(user, password string) (bool, error) {
	data, err := json.Marshal(map[string]string{
		"username": user,
		"password": password,
	})
	if err != nil {
		return false, err
	}
	req, err := http.NewRequest(http.MethodPost, a.config.Endpoint, bytes.NewReader(data))
	if err != nil {
		return false, err
	}
	if a.config.Headers != nil {
		for key, value := range a.config.Headers {
			req.Header.Add(key, value)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return false, errors.New(fmt.Sprintf("failed to send an auth request: %v", err))
	}
	return true, nil
}
