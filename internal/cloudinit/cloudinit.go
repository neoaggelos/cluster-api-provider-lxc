package cloudinit

import (
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"
)

// CloudInit represents a limited subset of cloud-config that is supported.
type CloudConfig struct {
	WriteFiles  []File   `json:"write_files"`
	RunCommands []string `json:"runcmd"`
}

type File struct {
	Path        string `json:"path"`
	Owner       string `json:"owner"`
	Permissions string `json:"permissions"`
	Content     string `json:"content"`
}

// Parse a cloud-init YAML manifest (only a limited subset of keys of the CloudConfig struct are allowed).
// If the manifest starts with `## template: jinja`, then an optional replacer is applied on the manifest before parsing.
func Parse(raw string, replacer *strings.Replacer) (CloudConfig, error) {
	raw, isJinja := strings.CutPrefix(raw, "## template: jinja\n")
	if isJinja && replacer != nil {
		raw = replacer.Replace(raw)
	}

	raw, isCloudConfig := strings.CutPrefix(raw, "#cloud-config\n")
	if !isCloudConfig {
		return CloudConfig{}, fmt.Errorf("missing required header #cloud-config")
	}

	var cloudConfig CloudConfig
	if err := yaml.UnmarshalStrict([]byte(raw), &cloudConfig, yaml.DisallowUnknownFields); err != nil {
		return CloudConfig{}, fmt.Errorf("failed parsing cloud-config YAML: %w", err)
	}

	return cloudConfig, nil
}
