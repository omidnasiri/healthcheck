package model

import (
	"errors"
	"net/http"

	"gorm.io/gorm"
)

type Endpoint struct {
	gorm.Model
	URL                string
	Interval           int // in seconds
	HTTPMethod         HTTPMethod
	HTTPRequestHeaders string
	HTTPRequestBody    string
	Retries            int // retries before submitting failure
	LastStatus         bool
	ActiveCheck        bool
	CheckLogs          []CheckLog
	Headers            map[string]string `gorm:"-:all"`
}

type HTTPMethod string

const (
	MethodGet     HTTPMethod = http.MethodGet
	MethodPut     HTTPMethod = http.MethodPut
	MethodPost    HTTPMethod = http.MethodPost
	MethodHead    HTTPMethod = http.MethodHead
	MethodPatch   HTTPMethod = http.MethodPatch
	MethodTrace   HTTPMethod = http.MethodTrace
	MethodDelete  HTTPMethod = http.MethodDelete
	MethodOptions HTTPMethod = http.MethodOptions
	MethodConnect HTTPMethod = http.MethodConnect
)

func (s HTTPMethod) Validate() error {
	if s != MethodGet && s != MethodPut && s != MethodPost &&
		s != MethodHead && s != MethodPatch && s != MethodTrace &&
		s != MethodDelete && s != MethodOptions && s != MethodConnect {
		return errors.New("invalid HTTP method")
	}
	return nil
}
