package lxc

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/lxc/incus/v6/shared/cliconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Configuration struct {
	// Server URL. Can be either `https://` or `unix://`
	ServerURL string `yaml:"server"`

	// Client certificate and key, only used if ServerURL is not "unix://..."
	ClientCrt          string `yaml:"client-crt"`
	ClientKey          string `yaml:"client-key"`
	ServerCrt          string `yaml:"server-crt"`
	InsecureSkipVerify bool   `yaml:"insecure-skip-verify"`

	// Project name
	Project string `yaml:"project"`
}

// ConfigurationFromKubernetesSecret parses a Kubernetes secret and derives Configuration for connecting to Incus.
//
// The secret can be created like this:
//
//	```bash
//
//	# create a client certificate and key trusted by incus
//	$ incus remote generate-certificate
//	$ sudo incus config trust add-certificate ~/.config/incus/client.crt
//
//	# [manually] generate kubernetes secret
//	$ kubectl create secret generic incus-secret \
//		--from-literal=server="https://10.0.0.49:8443" \
//		--from-literal=server-crt="$(sudo cat /var/lib/incus/cluster.crt)" \
//		--from-literal=client-crt="$(cat ~/.config/incus/client.crt)" \
//		--from-literal=client-key="$(cat ~/.config/incus/client.key)" \
//		--from-literal=project="default"
//
//	# [manually] generate kubernetes secret with insecure skip verify
//	$ kubectl create secret generic lxd-secret \
//		--from-literal=server=https://10.0.1.2:8901 \
//		--from-literal=insecure-skip-verify=true \
//		--from-literal=client-crt="$(cat ~/.config/incus/client.crt)" \
//		--from-literal=client-key="$(cat ~/.config/incus/client.key)" \
//		--from-literal=project="default"
//
//	```
func ConfigurationFromKubernetesSecret(secret *corev1.Secret) Configuration {
	insecureSkipVerify, _ := strconv.ParseBool(string(secret.Data["insecure-skip-verify"]))
	return Configuration{
		ServerURL:          string(secret.Data["server"]),
		Project:            string(secret.Data["project"]),
		ClientCrt:          string(secret.Data["client-crt"]),
		ClientKey:          string(secret.Data["client-key"]),
		ServerCrt:          string(secret.Data["server-crt"]),
		InsecureSkipVerify: insecureSkipVerify,
	}
}

// ToSecret generates secret data from a Configuration struct.
func (o Configuration) ToSecret(name string, namespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"clusterctl.cluster.x-k8s.io/move": "true",
			},
		},
		Data: map[string][]byte{
			"server":               []byte(o.ServerURL),
			"project":              []byte(o.Project),
			"client-crt":           []byte(o.ClientCrt),
			"client-key":           []byte(o.ClientKey),
			"server-crt":           []byte(o.ServerCrt),
			"insecure-skip-verify": []byte(fmt.Sprintf("%t", o.InsecureSkipVerify)),
		},
	}
}

// ConfigurationFromLocal attempts to load client options from the local node configuration file.
// ConfigurationFromLocal will attempt to use well-known locations.
// ConfigurationFromLocal returns the loaded Configuration, as well as the path of the config file.
func ConfigurationFromLocal(configFile string, forceRemoteName string, requireHTTPS bool) (Configuration, string, error) {
	var tryConfigFiles []string
	if configFile == "" {
		tryConfigFiles = getDefaultConfigFiles()
	} else {
		tryConfigFiles = []string{configFile}
	}

	var errs []error
	for _, configFile := range tryConfigFiles {
		config, err := cliconfig.LoadConfig(configFile)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to load credentials from %q: %w", configFile, err))
			continue
		}

		remoteName := forceRemoteName
		if remoteName == "" {
			remoteName = config.DefaultRemote
		}

		remote, ok := config.Remotes[remoteName]
		if !ok {
			errs = append(errs, fmt.Errorf("failed to load credentials from %q: remote %q not found", configFile, remoteName))
			continue
		}

		if requireHTTPS && !strings.HasPrefix(remote.Addr, "https://") {
			errs = append(errs, fmt.Errorf("failed to load credentials from %q: remote address %q must use HTTPS", configFile, remote.Addr))
			continue
		}

		if strings.HasPrefix(remote.Addr, "unix://") || strings.HasPrefix(remote.Addr, "http://") {
			return Configuration{
				ServerURL: remote.Addr,
				Project:   remote.Project,
			}, configFile, nil
		}

		if !config.HasClientCertificate() {
			errs = append(errs, fmt.Errorf("failed to load credentials from %q: no client certificate", configFile))
			continue
		}

		serverCrt, err := os.ReadFile(config.ServerCertPath(remoteName))
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to load credentials from %q: cannot read server certificate for remote %q: %v", configFile, remoteName, err))
			continue
		}

		clientCrt, err := os.ReadFile(config.ConfigPath("client.crt"))
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to load credentials from %q: cannot read client certificate for remote %q: %v", configFile, remoteName, err))
			continue
		}

		clientKey, err := os.ReadFile(config.ConfigPath("client.key"))
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to load credentials from %q: cannot read client key for remote %q: %v", configFile, remoteName, err))
			continue
		}

		return Configuration{
			ServerURL: remote.Addr,
			ServerCrt: string(serverCrt),
			ClientCrt: string(clientCrt),
			ClientKey: string(clientKey),
			Project:   remote.Project,
		}, configFile, nil
	}

	return Configuration{}, "", errors.Join(errs...)
}
