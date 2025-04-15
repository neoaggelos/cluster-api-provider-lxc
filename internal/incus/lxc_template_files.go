package incus

import (
	"bytes"
	"context"
	"fmt"

	"github.com/lxc/cluster-api-provider-incus/internal/static"
	"github.com/lxc/incus/v6/shared/api"
)

func (c *Client) ensureInstanceTemplateFiles(ctx context.Context, instanceName string) error {
	metadata, _, err := c.Client.GetInstanceMetadata(instanceName)
	if err != nil {
		return fmt.Errorf("failed to GetInstanceMetadata: %w", err)
	}

	var mustUpdateMetadata bool
	for _, file := range []struct {
		templateName string
		content      string
		filePath     string
		mode         string
	}{
		{templateName: "capn-install-kubeadm.tpl", filePath: "/opt/cluster-api/install-kubeadm.sh", content: static.InstallKubeadmScript(), mode: "0755"},
		{templateName: "capn-kube-flannel.tpl", filePath: "/opt/cluster-api/kube-flannel.yaml", content: static.KubeFlannelTemplate(), mode: "0644"},
		{templateName: "capn-kube-proxy-config.tpl", filePath: "/opt/cluster-api/kube-proxy-config-lxc.yaml", content: static.KubeProxyConfigTemplate(), mode: "0644"},
		{templateName: "capn-kube-vip.tpl", filePath: "/opt/cluster-api/kube-vip-pod.yaml", content: static.KubeVIPTemplate(), mode: "0644"},
	} {
		if _, ok := metadata.Templates[file.filePath]; !ok {
			if err := c.Client.CreateInstanceTemplateFile(instanceName, file.templateName, bytes.NewReader([]byte(file.content))); err != nil {
				// TODO: do not fail if already exists
				return fmt.Errorf("failed to CreateInstanceTemplateFile(%q): %w", file.templateName, err)
			}

			metadata.Templates[file.filePath] = &api.ImageMetadataTemplate{
				When:       []string{"create", "copy", "start"},
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
