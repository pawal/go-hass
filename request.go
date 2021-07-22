package hass

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Doer represents an http client that can "Do" a request
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Access is the access and credentials for the API
type Access struct {
	host        string
	password    string
	token       string
	bearertoken string
	client      Doer
}

// NewAccess returns a new *Access to be used to interface with the
// Home Assistant system.
func NewAccess(host, password string) *Access {
	return &Access{
		host:     host,
		password: password,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// SetAccess changes login credentials for API access
func (a *Access) SetAccess(host, password string) {
	a.host = host
	a.password = password
}

// SetClient allows you to specify a different http client than the default
func (a *Access) SetClient(client Doer) {
	a.client = client
}

// SetToken sets the X-HASSIO-KEY header
func (a *Access) SetToken(token string) {
	a.token = token
}

// SetBearerToken sets the Authentiation: Bearer header
// Long Lived Access Tokens can be generated from the HASS UI
func (a *Access) SetBearerToken(token string) {
	a.bearertoken = "Bearer " + token
}

func (a *Access) httpGet(path string, v interface{}) error {
	req, err := http.NewRequest("GET", a.host+path, nil)
	println(a.host + path)

	if err != nil {
		return err
	}

	if a.password != "" {
		req.Header.Set("x-ha-access", a.password)
	}

	if a.token != "" {
		req.Header.Set("X-HASSIO-KEY", a.token)
	}

	if a.bearertoken != "" {
		req.Header.Set("Authorization", a.bearertoken)
	}

	success := false
	for i := 0; i < 3; i++ {
		func() {
			var resp *http.Response
			resp, err = a.client.Do(req)
			if err != nil {
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				err = errors.New("hass: status not OK: " + resp.Status)
				return
			}

			dec := json.NewDecoder(resp.Body)
			err = dec.Decode(v)
			success = true
		}()

		if success {
			break
		}
	}

	return err
}

func (a *Access) httpPost(path string, v interface{}) error {
	var req *http.Request

	if v != nil {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}

		req, err = http.NewRequest("POST", a.host+path, bytes.NewReader(data))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
	} else {
		var err error
		req, err = http.NewRequest("POST", a.host+path, nil)
		if err != nil {
			return err
		}
	}

	if a.password != "" {
		req.Header.Set("x-ha-access", a.password)
	}

	if a.token != "" {
		req.Header.Set("X-HASSIO-KEY", a.token)
	}

	if a.bearertoken != "" {
		req.Header.Set("Authorization", a.bearertoken)
	}

	var err error
	success := false
	for i := 0; i < 3; i++ {
		func() {
			var resp *http.Response
			resp, err = a.client.Do(req)
			if err != nil {
				return
			}

			defer resp.Body.Close()

			success = true
		}()

		if success {
			break
		}
	}

	return err
}
