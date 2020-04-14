package middleware

import (
	"github.com/amimof/multikube/pkg/proxy"
	"net/http"
)

// Middleware represents a multikube middleware
type Middleware func(*proxy.Proxy, http.Handler) http.Handler
