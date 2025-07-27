package lxc_test

import (
	"fmt"
	"testing"

	"github.com/lxc/incus/v6/shared/api"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"

	. "github.com/onsi/gomega"
)

func simplestreamsImage(server string, name string) api.InstanceSource {
	return api.InstanceSource{
		Type:     "image",
		Protocol: "simplestreams",
		Server:   server,
		Alias:    name,
	}
}

func TestTryParseImageSource(t *testing.T) {
	for _, tc := range []struct {
		server            string
		image             string
		expectErr         bool
		expectParsed      bool
		expectImageSource api.InstanceSource
	}{
		// verify incus prefixes
		{server: "incus", image: "image-name"},
		{server: "incus", image: "unknown:image", expectErr: true},
		{server: "incus", image: "ubuntu:24.04", expectParsed: true, expectImageSource: simplestreamsImage("https://images.linuxcontainers.org", "ubuntu/24.04/cloud")},
		{server: "incus", image: "debian:12", expectParsed: true, expectImageSource: simplestreamsImage("https://images.linuxcontainers.org", "debian/12/cloud")},
		{server: "incus", image: "images:almalinux/9/cloud", expectParsed: true, expectImageSource: simplestreamsImage("https://images.linuxcontainers.org", "almalinux/9/cloud")},
		{server: "incus", image: "capi:kubeadm/v1.33.0", expectParsed: true, expectImageSource: simplestreamsImage("https://d14dnvi2l3tc5t.cloudfront.net", "kubeadm/v1.33.0")},
		{server: "incus", image: "capi-stg:kubeadm/v1.33.0", expectParsed: true, expectImageSource: simplestreamsImage("https://djapqxqu5n2qu.cloudfront.net", "kubeadm/v1.33.0")},
		// verify lxd prefixes
		{server: "lxd", image: "image-name"},
		{server: "lxd", image: "unknown:image", expectErr: true},
		{server: "lxd", image: "ubuntu:24.04", expectParsed: true, expectImageSource: simplestreamsImage("https://cloud-images.ubuntu.com/releases/", "24.04")},
		{server: "lxd", image: "debian:12", expectParsed: true, expectImageSource: simplestreamsImage("https://images.lxd.canonical.com", "debian/12/cloud")},
		{server: "lxd", image: "images:almalinux/9/cloud", expectParsed: true, expectImageSource: simplestreamsImage("https://images.lxd.canonical.com", "almalinux/9/cloud")},
		{server: "lxd", image: "capi:kubeadm/v1.33.0", expectParsed: true, expectImageSource: simplestreamsImage("https://d14dnvi2l3tc5t.cloudfront.net", "kubeadm/v1.33.0")},
		{server: "lxd", image: "capi-stg:kubeadm/v1.33.0", expectParsed: true, expectImageSource: simplestreamsImage("https://djapqxqu5n2qu.cloudfront.net", "kubeadm/v1.33.0")},
		// verify prefixes for unknown
		{server: "unknown", image: "image-name"},
		{server: "unknown", image: "ubuntu:24.04", expectErr: true},
	} {
		t.Run(fmt.Sprintf("%s/%s", tc.server, tc.image), func(t *testing.T) {
			g := NewWithT(t)

			image, parsed, err := lxc.TryParseImageSource(tc.server, tc.image)
			if tc.expectErr {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).ToNot(HaveOccurred())
			if !tc.expectParsed {
				g.Expect(parsed).To(BeFalse())
				return
			}

			g.Expect(image).To(Equal(tc.expectImageSource))
		})
	}
}
