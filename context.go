package multikube

import (
	"net/http"
)

type Context struct {
	Request  *http.Request
	Response *http.ResponseWriter
	APIServer *APIServer
}