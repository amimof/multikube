package middleware

import (
	"net/http"
	"github.com/amimof/multikube/pkg/proxy"
)

// Middleware represents a multikube middleware
type Middleware func(*proxy.Proxy, http.Handler) http.Handler