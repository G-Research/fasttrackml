package config

import (
	"fmt"
	"strings"
)

// NormaliseListenAddress normalise listenAddress parameter.
func NormaliseListenAddress(listenAddress string) string {
	if strings.Contains(listenAddress, "http://") || strings.Contains(listenAddress, "https://") {
		return listenAddress
	}
	return fmt.Sprintf("http://%s", listenAddress)
}
