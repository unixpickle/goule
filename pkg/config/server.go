package config

type ServerSettings struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
}
