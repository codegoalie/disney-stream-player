package utils

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"
)

// DefaultHTTPClient defines a nicely configured HTTP client
var DefaultHTTPClient = &http.Client{
	Timeout: time.Second * 30,
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Second * 10,
		}).Dial,
		TLSHandshakeTimeout: time.Second * 10,
	},
}

// HTTPGet issues an HTTP GET request to URL and returns the body in a buffer
func HTTPGet(url string) (*bytes.Buffer, error) {
	resp, err := DefaultHTTPClient.Get(url)
	if err != nil {
		err = fmt.Errorf("failed to issue HttpGet: %w", err)
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		err = fmt.Errorf("failed to read HttpGet body: %w", err)
		return nil, err
	}

	return buf, nil
}
