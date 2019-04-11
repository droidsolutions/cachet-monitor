package cachet

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
)

type CachetAPI struct {
	URL      string `json:"url"`
	Token    string `json:"token"`
	Insecure bool   `json:"insecure"`
}

type CachetResponse struct {
	Data json.RawMessage `json:"data"`
}

// TODO: test
func (api CachetAPI) Ping() error {
	resp, _, err := api.NewRequest("GET", "/ping", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("API Responded with non-200 status code")
	}

	defer resp.Body.Close()

	return nil
}

// SendMetric adds a data point to a cachet monitor
func (api CachetAPI) SendMetric(id int, lag int64) {
	logrus.Debugf("Sending lag metric ID:%d RTT %vms", id, lag)

	jsonBytes, _ := json.Marshal(map[string]interface{}{
		"value":     lag,
		"timestamp": time.Now().Unix(),
	})

	resp, _, err := api.NewRequest("POST", "/metrics/"+strconv.Itoa(id)+"/points", jsonBytes)
	if err != nil || resp.StatusCode != 200 {
		logrus.Warnf("Could not log metric! ID: %d, err: %v", id, err)
	}

	defer resp.Body.Close()
}

// NewRequest wraps http.NewRequest
func (api CachetAPI) NewRequest(requestType, url string, reqBody []byte) (*http.Response, CachetResponse, error) {
	req, err := http.NewRequest(requestType, api.URL+url, bytes.NewBuffer(reqBody))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cachet-Token", api.Token)

	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: api.Insecure}
	client := &http.Client{
		Transport: transport,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, CachetResponse{}, err
	}

	var body struct {
		Data json.RawMessage `json:"data"`
	}
	err = json.NewDecoder(res.Body).Decode(&body)

	defer req.Body.Close()

	return res, body, err
}

// GetComponentStatus
func (api CachetAPI) GetComponentStatus(componentID int) (int, error) {
	resp, body, err := api.NewRequest("GET", "/components/"+strconv.Itoa(componentID), nil)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("Invalid status code. Received %d", resp.StatusCode)
	}

	data := IncidentResponse{}

	if err := json.Unmarshal(body.Data, &data); err != nil {
		return 0, fmt.Errorf("Cannot parse component body: %v. Err = %v", string(body.Data), err)
	}

	return int(data.Status), nil
}
