package incus

import (
	"context"
	"fmt"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/tls"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Client struct {
	Client incus.InstanceServer
}

func New(ctx context.Context, opts Options) (*Client, error) {
	log := log.FromContext(ctx).WithValues("lxc.server", opts.ServerURL)

	switch {
	case opts.InsecureSkipVerify:
		log = log.WithValues("lxc.insecure-skip-verify", true)
		opts.ServerCrt = ""
	case opts.ServerCrt == "":
		log = log.WithValues("lxc.server-crt", "<unset>")
	case opts.ServerCrt != "":
		if fingerprint, err := tls.CertFingerprintStr(opts.ServerCrt); err == nil && len(fingerprint) >= 12 {
			log = log.WithValues("lxc.server-crt", fingerprint[:12])
		}
	}

	if fingerprint, err := tls.CertFingerprintStr(opts.ClientCrt); err == nil && len(fingerprint) >= 12 {
		log = log.WithValues("lxc.client-crt", fingerprint[:12])
	}

	client, err := incus.ConnectIncusWithContext(ctx, opts.ServerURL, &incus.ConnectionArgs{
		TLSServerCert:      opts.ServerCrt,
		TLSClientCert:      opts.ClientCrt,
		TLSClientKey:       opts.ClientKey,
		InsecureSkipVerify: opts.InsecureSkipVerify,
		SkipGetServer:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize incus client: %w", err)
	}

	if opts.Project != "" {
		log = log.WithValues("lxc.project", opts.Project)
		client = client.UseProject(opts.Project)
	}

	log.V(2).Info("Initialized new client")

	return &Client{Client: client}, nil
}
