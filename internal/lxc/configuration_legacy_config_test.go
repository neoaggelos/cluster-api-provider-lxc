package lxc_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"

	. "github.com/onsi/gomega"
)

// TestConfigurationFromLocalLegacyAddr verifies that client configuration files
// written with the pre-v7 single `addr:` field still parse correctly after the
// incus/v7 bump, where cliconfig.Remote.Addr (string) became Addrs ([]string).
// incus v7 keeps reading the legacy `addr:` key via Remote.UnmarshalYAML, so old
// config files remain valid.
func TestConfigurationFromLocalLegacyAddr(t *testing.T) {
	t.Run("unix remote", func(t *testing.T) {
		g := NewWithT(t)

		dir := t.TempDir()
		writeFile(g, filepath.Join(dir, "config.yml"), `default-remote: node
remotes:
  node:
    addr: unix:///var/lib/incus/unix.socket
    public: false
    protocol: incus
`)

		cfg, _, err := lxc.ConfigurationFromLocal(filepath.Join(dir, "config.yml"), "", false)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(cfg.ServerURL).To(Equal("unix:///var/lib/incus/unix.socket"))
	})

	t.Run("https remote with certificates", func(t *testing.T) {
		g := NewWithT(t)

		dir := t.TempDir()
		writeFile(g, filepath.Join(dir, "config.yml"), `default-remote: cluster
remotes:
  cluster:
    addr: https://10.0.0.49:8443
    public: false
    protocol: incus
    project: default
`)
		writeFile(g, filepath.Join(dir, "client.crt"), "client-cert")
		writeFile(g, filepath.Join(dir, "client.key"), "client-key")
		writeFile(g, filepath.Join(dir, "servercerts", "cluster.crt"), "server-cert")

		cfg, _, err := lxc.ConfigurationFromLocal(filepath.Join(dir, "config.yml"), "", true)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(cfg.ServerURL).To(Equal("https://10.0.0.49:8443"))
		g.Expect(cfg.Project).To(Equal("default"))
		g.Expect(cfg.ServerCrt).To(Equal("server-cert"))
		g.Expect(cfg.ClientCrt).To(Equal("client-cert"))
		g.Expect(cfg.ClientKey).To(Equal("client-key"))
	})
}

func writeFile(g *WithT, path, content string) {
	g.Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
	g.Expect(os.WriteFile(path, []byte(content), 0o600)).To(Succeed())
}
