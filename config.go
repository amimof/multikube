package multikube

type Config struct {
	LogPath    string       `json:"logPath,omitempty"`
	APIServers []*APIServer `json:"apiServers,omitempty"`
}
