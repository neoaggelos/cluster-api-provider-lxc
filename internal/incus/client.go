package incus

import (
	"context"
	"fmt"
	"strconv"

	incus "github.com/lxc/incus/v6/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
)

type Client struct {
	Client incus.InstanceServer
}

type Options struct {
	ServerURL string `json:"serverURL"`

	Project string `json:"project"`

	ClientCert string `json:"clientCrt"`
	ClientKey  string `json:"clientKey"`

	ServerCert         string `json:"serverCrt"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
}

// NewOptionsFromSecret parses a Kubernetes secret and derives Options for connecting to Incus.
//
// The secret can be created like this:
//
// ```bash
//
//	# create a client certificate and key trusted by incus
//	$ incus remote generate-certificate
//	$ sudo incus config trust add-certificate ~/.config/incus/client.crt
//
//	# generate kubernetes secret
//	$ kubectl create secret generic incus-secret \
//		--from-literal=server="https://10.0.0.49:8443" \
//		--from-literal=server-crt="$(sudo cat /var/lib/incus/cluster.crt)" \
//		--from-literal=client-crt="$(cat ~/.config/incus/client.crt)" \
//		--from-literal=client-key="$(cat ~/.config/incus/client.key)" \
//		--from-literal=project="default"
//
//	# or with insecure skip verify
//	$ kubectl create secret generic lxd-secret \
//		--from-literal=server=https://10.0.1.2:8901 \
//		--from-literal=insecure-skip-verify=true \
//		--from-literal=client-crt="$(cat ~/.config/incus/client.crt)" \
//		--from-literal=client-key="$(cat ~/.config/incus/client.key)" \
//		--from-literal=project="default"
//
// ```
func NewOptionsFromSecret(secret *corev1.Secret) Options {
	insecureSkipVerify, _ := strconv.ParseBool(string(secret.Data["insecure-skip-verify"]))
	return Options{
		ServerURL:          string(secret.Data["server"]),
		Project:            string(secret.Data["project"]),
		ClientCert:         string(secret.Data["client-crt"]),
		ClientKey:          string(secret.Data["client-key"]),
		ServerCert:         string(secret.Data["server-crt"]),
		InsecureSkipVerify: insecureSkipVerify,
	}
}

func New(ctx context.Context, opts Options) (*Client, error) {
	log := log.FromContext(ctx).WithValues("server", opts.ServerURL)

	client, err := incus.ConnectIncusWithContext(ctx, opts.ServerURL, &incus.ConnectionArgs{
		TLSServerCert:      opts.ServerCert,
		TLSClientCert:      opts.ClientCert,
		TLSClientKey:       opts.ClientKey,
		InsecureSkipVerify: opts.InsecureSkipVerify,
		SkipGetServer:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize incus client: %w", err)
	}

	if opts.Project != "" {
		log = log.WithValues("project", opts.Project)
		client = client.UseProject(opts.Project)
	}

	log.V(4).Info("Initialized new incus client")

	return &Client{Client: client}, nil
}
