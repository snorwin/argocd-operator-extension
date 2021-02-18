package helm

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// Client is a Helm client
type Client interface {
	Install(release string, chart *chart.Chart, values chartutil.Values) error
	Upgrade(release string, chart *chart.Chart, values chartutil.Values, install bool) error
	Uninstall(release string) error
}

// ClientFactory provides an abstraction to create a new namespaced Client
type ClientFactory func(namespace string, options ...ClientOption) (Client, error)

// client implements the Client interface
type client struct {
	action.Configuration

	maxHistory int
	namespace  string
	driver     string
	logger     func(format string, v ...interface{})
}

// NewClientForNamespace is a ClientFactory
func NewClientForNamespace(namespace string, options ...ClientOption) (Client, error) {
	c := &client{namespace: namespace}

	// apply the ClientOptions to the client
	for _, option := range options {
		option(c)
	}

	// initialize the client configuration
	if err := c.Init(cli.New().RESTClientGetter(), c.namespace, c.driver, c.logger); err != nil {
		return nil, err
	}

	return c, nil
}

// Install creates a action.Install with the client configuration and installs the given Helm chart and values
func (c *client) Install(release string, chart *chart.Chart, values chartutil.Values) error {
	install := action.NewInstall(&c.Configuration)
	install.ReleaseName = release
	install.Namespace = c.namespace
	if _, err := install.Run(chart, values); err != nil {
		return err
	}
	return nil
}

// Upgrade creates a action.Upgrade with the client configuration and upgrades or installs (if the install flag is set) the given Helm chart and values
func (c *client) Upgrade(release string, chart *chart.Chart, values chartutil.Values, install bool) error {
	if _, err := action.NewStatus(&c.Configuration).Run(release); err == driver.ErrReleaseNotFound && install {
		return c.Install(release, chart, values)
	} else if err == nil {
		upgrade := action.NewUpgrade(&c.Configuration)
		upgrade.MaxHistory = c.maxHistory
		if _, err = upgrade.Run(release, chart, values); err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

// Uninstall creates a action.Uninstall with the client configuration and uninstalls the given release
func (c *client) Uninstall(release string) error {
	if _, err := action.NewUninstall(&c.Configuration).Run(release); err != nil {
		if err == driver.ErrReleaseNotFound {
			err = nil
		}
		return err
	}
	return nil
}
