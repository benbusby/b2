package b2

import (
	"errors"
	"net/http"
	"time"
)

const APIPrefix string = "b2api/v2"

var B2Client = &http.Client{Timeout: 10 * time.Second}
var B2Error = errors.New("b2 client error")
