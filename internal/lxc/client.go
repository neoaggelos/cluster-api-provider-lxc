package lxc

import (
	"context"
	"fmt"
	"strings"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"github.com/lxc/incus/v6/shared/tls"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Client struct {
	incus.InstanceServer

	serverInfo *api.Server

	progressHandler func(api.Operation)
}

type Option func(c *Client)

// WithProgressHandler sets a custom progress handler for ongoing operations.
func (c *Client) WithProgressHandler(f func(api.Operation)) {
	c.progressHandler = f
}

func New(ctx context.Context, config Configuration, options ...Option) (*Client, error) {
	log := log.FromContext(ctx)

	var (
		client incus.InstanceServer
		err    error
	)
	switch {
	case strings.HasPrefix(config.ServerURL, "https://"):
		log = log.WithValues("lxc.server", config.ServerURL)
		switch {
		case config.InsecureSkipVerify:
			log = log.WithValues("lxc.insecure-skip-verify", true)
			config.ServerCrt = ""
		case config.ServerCrt == "":
			log = log.WithValues("lxc.server-crt", "<unset>")
		case config.ServerCrt != "":
			if fingerprint, err := tls.CertFingerprintStr(config.ServerCrt); err == nil && len(fingerprint) >= 12 {
				log = log.WithValues("lxc.server-crt", fingerprint[:12])
			}
		}

		if fingerprint, err := tls.CertFingerprintStr(config.ClientCrt); err == nil && len(fingerprint) >= 12 {
			log = log.WithValues("lxc.client-crt", fingerprint[:12])
		}

		if client, err = incus.ConnectIncusWithContext(ctx, config.ServerURL, &incus.ConnectionArgs{
			TLSServerCert:      config.ServerCrt,
			TLSClientCert:      config.ClientCrt,
			TLSClientKey:       config.ClientKey,
			InsecureSkipVerify: config.InsecureSkipVerify,
		}); err != nil {
			return nil, fmt.Errorf("failed to initialize client over HTTPS: %w", err)
		}
	case strings.HasPrefix(config.ServerURL, "unix://"):
		socket, ok := strings.CutPrefix(config.ServerURL, "unix://")
		if ok && socket == "" {
			if socket, err = findDefaultUnixSocketPath(); err != nil {
				return nil, fmt.Errorf("failed to detect default local unix socket path: %w", err)
			}
		}
		log = log.WithValues("lxc.server", fmt.Sprintf("unix://%s", socket))
		if client, err = incus.ConnectIncusUnixWithContext(ctx, socket, nil); err != nil {
			return nil, fmt.Errorf("failed to initialize client over unix socket: %w", err)
		}
	default:
		return nil, fmt.Errorf("server %q is not unix:// or https://", config.ServerURL)
	}

	if config.Project != "" {
		log = log.WithValues("lxc.project", config.Project)
		client = client.UseProject(config.Project)
	}

	server, _, err := client.GetServer()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve server information: %w", err)
	}

	log.V(5).Info("Initialized client")

	c := &Client{InstanceServer: client, serverInfo: server}
	for _, o := range options {
		o(c)
	}

	return c, nil
}

// WithTarget returns a copy of the client and a set target host.
// WithTarget will ignore the target argument if server is not clustered.
func (c *Client) WithTarget(target string) *Client {
	if c.SupportsInstanceTarget() != nil {
		return c
	}
	return &Client{
		InstanceServer:  c.UseTarget(target),
		serverInfo:      c.serverInfo,
		progressHandler: c.progressHandler,
	}
}
