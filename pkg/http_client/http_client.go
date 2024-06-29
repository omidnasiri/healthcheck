package httpclient

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"time"
)

func Do(
	ctx context.Context,
	method, url string,
	data []byte,
	timeout time.Duration,
	headers map[string]string,
) (
	body []byte,
	httpStatusCode int,
	err error,
) {
	transport := &http.Transport{
		DisableKeepAlives: true,
		Dial: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: -1,
		}).Dial,
		// TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: timeout,
	}

	var client = &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	var req *http.Request
	if data != nil {
		buf := bytes.NewBuffer(data)
		req, err = http.NewRequest(
			method,
			url,
			buf,
		)
	} else {
		req, err = http.NewRequest(
			method,
			url,
			http.NoBody,
		)
	}
	if err != nil {
		return nil, -1, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, -1, err
	}
	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, -1, err
	}

	return body, res.StatusCode, nil
}
