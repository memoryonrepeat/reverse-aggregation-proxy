package config

import (
	"time"
)

var MaxTop = 10
var DefaultTop = 5
var DefaultSkip = 0
var BaseURL = "https://s3-eu-west-1.amazonaws.com/test-golang-recipes/"
var RequestTimeout time.Duration = 2
var ClientTimeout time.Duration = 10
