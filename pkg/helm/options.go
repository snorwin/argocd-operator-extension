package helm

import (
	"fmt"

	"github.com/go-logr/logr"
)

// ClientOption defines a function types to apply options to the client configuration
type ClientOption func(*client)

// WithHelmDriver set the storage helm driver (configmap`, `secret`, `memory`) of the client configuration
func WithHelmDriver(driver string) ClientOption {
	return func(client *client) {
		client.driver = driver
	}
}

// WithLogger injects a logr.Logger to the client configuration
func WithLogger(logger logr.Logger) ClientOption {
	return func(c *client) {
		c.logger = func(format string, v ...interface{}) {
			logger.V(4).Info(fmt.Sprintf(format, v...))
		}
	}
}
