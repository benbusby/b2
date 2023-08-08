package utils

import (
	"errors"
	"net/http"
	"time"
)

const APIPrefix string = "b2api/v2"

var Client = &http.Client{Timeout: 10 * time.Second}
var Error = errors.New("b2 client error")
