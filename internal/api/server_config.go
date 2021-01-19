package api

// ServerConfig contains the configurations for a Server
type ServerConfig struct {
	Port         int
	CertFilePath string
	KeyFilePath  string
}
