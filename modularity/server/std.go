package server

import (
	"net/http"

	"github.com/onemoreteam/httpframework"
)

var Default = httpframework.NewServer(&http.Server{})
