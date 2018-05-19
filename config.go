package multikube

type Config struct {
	Name string `json:"name,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	CA string `json:"ca,omitempty"`
	Cert string `json:"cert,omitempty"`
	Key string `json:"key,omitempty"`
}