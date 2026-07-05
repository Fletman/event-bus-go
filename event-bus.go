package main

import (
	"os"

	"event-bus.go/api"
)

func env(env_key string, default_val string) string {
	val := os.Getenv(env_key)
	if val == "" {
		return default_val
	}
	return val
}

func main() {
	port := env("SERVER_PORT", "8080")
	event_bus_config := env("EVENT_BUS", "channel")

	// /etc/cert-manager/tls.*
	var tls_config *api.TlsConfig
	cert_path := env("TLS_CERT_PATH", "")
	key_path := env("TLS_KEY_PATH", "")
	if cert_path == "" || key_path == "" {
		tls_config = nil
	} else {
		tls_config = &api.TlsConfig{
			CertPath: cert_path,
			KeyPath:  key_path,
		}
	}

	api_server, err := api.NewServer(event_bus_config)
	if err != nil {
		panic(err)
	}
	err = api_server.StartServer(port, tls_config)
	if err != nil {
		panic(err)
	}
}
