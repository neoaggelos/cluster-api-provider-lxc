package incus

import (
	"bytes"
	"fmt"

	"github.com/lxc/incus/v6/shared/api"

	"github.com/lxc/cluster-api-provider-incus/internal/static"
)

func (c *Client) ensureInstanceTemplateFiles(instanceName string, isControlPlaneMachine bool) error {
	metadata, _, err := c.Client.GetInstanceMetadata(instanceName)
	if err != nil {
		return fmt.Errorf("failed to GetInstanceMetadata: %w", err)
	}

	var mustUpdateMetadata bool
	for _, file := range []struct {
		templateName string
		content      string
		filePath     string
		guard        bool
	}{
		// inject install-kubeadm.sh in all nodes
		{templateName: "capn-install-kubeadm.tpl", guard: true, filePath: "/opt/cluster-api/install-kubeadm.sh", content: static.InstallKubeadmScript()},
		// add kube-proxy-config.yaml, kube-flannel and kube-vip manifests only on control plane machines
		{templateName: "capn-kube-proxy-config.tpl", guard: isControlPlaneMachine, filePath: "/opt/cluster-api/kube-proxy-config-lxc.yaml", content: static.KubeProxyConfigTemplate()},
	} {
		if !file.guard {
			continue
		}

		if _, ok := metadata.Templates[file.filePath]; !ok {
			if err := c.Client.CreateInstanceTemplateFile(instanceName, file.templateName, bytes.NewReader([]byte(file.content))); err != nil {
				// TODO: do not fail if already exists
				return fmt.Errorf("failed to CreateInstanceTemplateFile(%q): %w", file.templateName, err)
			}

			metadata.Templates[file.filePath] = &api.ImageMetadataTemplate{
				When:       []string{"create"},
				CreateOnly: true,
				Template:   file.templateName,
			}

			mustUpdateMetadata = true
		}
	}

	if mustUpdateMetadata {
		if err := c.Client.UpdateInstanceMetadata(instanceName, *metadata, ""); err != nil {
			return fmt.Errorf("failed to UpdateInstanceMetadata: %w", err)
		}
	}

	return nil
}
