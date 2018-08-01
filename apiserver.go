package multikube

type ResourceSpec struct {
	ApiVersion string
	Name       string
	Namespace  string
	Type       interface{}
	Path       string
}

type apiserver struct {
	Name     string `json:"hostname,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	CA       string `json:"ca,omitempty"`
	Cert     string `json:"cert,omitempty"`
	Key      string `json:"key,omitempty"`
	Insecure bool   `json:"insecure,omitempty"`
	cache    *Cache
}

type APIServer struct {
	apiserver
}

func (c *APIServer) Hostname() string {
	return c.apiserver.Hostname
}

func (c *APIServer) CA() string {
	return c.apiserver.CA
}

func (c *APIServer) Cert() string {
	return c.apiserver.Cert
}

func (c *APIServer) Key() string {
	return c.apiserver.Key
}

func (c *APIServer) Insecure() bool {
	return c.apiserver.Insecure
}

// Cache returns the current cache instance of the cluster
func (c *APIServer) Cache() *Cache {
	if c.cache == nil {
		c.cache = NewCache()
	}
	return c.cache
}
