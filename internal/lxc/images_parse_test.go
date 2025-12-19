package lxc_test

import (
	"fmt"
	"testing"

	"github.com/lxc/incus/v6/shared/api"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"

	. "github.com/onsi/gomega"
)

func simplestreamsImage(server string, name string) api.ImageSource {
	return api.ImageSource{
		Protocol: "simplestreams",
		Server:   server,
		Alias:    name,
	}
}

func ociImage(server string, name string) api.ImageSource {
	return api.ImageSource{
		Protocol: "oci",
		Server:   server,
		Alias:    name,
	}
}

func TestParseImage(t *testing.T) {
	for _, tc := range []struct {
		server            string
		image             string
		expectParseErr    bool
		expectParse       bool
		expectErr         bool
		expectImageSource api.ImageSource
	}{
		// verify incus prefixes
		{server: "incus", image: "image-name"},
		{server: "incus", image: "unknown:image", expectParseErr: true},
		{server: "incus", image: "ubuntu:24.04", expectParse: true, expectImageSource: simplestreamsImage("https://images.linuxcontainers.org", "ubuntu/24.04/cloud")},
		{server: "incus", image: "debian:12", expectParse: true, expectImageSource: simplestreamsImage("https://images.linuxcontainers.org", "debian/12/cloud")},
		{server: "incus", image: "images:almalinux/9/cloud", expectParse: true, expectImageSource: simplestreamsImage("https://images.linuxcontainers.org", "almalinux/9/cloud")},
		{server: "incus", image: "capi:kubeadm/v1.33.0", expectParse: true, expectImageSource: simplestreamsImage("https://images.capn.open-cloud.xyz", "kubeadm/v1.33.0")},
		{server: "incus", image: "capi-stg:kubeadm/v1.33.0", expectParse: true, expectImageSource: simplestreamsImage("https://images-stg.capn.open-cloud.xyz", "kubeadm/v1.33.0")},
		{server: "incus", image: "kind:v1.33.0", expectParse: true, expectImageSource: ociImage("https://docker.io", "kindest/node:v1.33.0")},
		// verify lxd prefixes
		{server: "lxd", image: "image-name"},
		{server: "lxd", image: "unknown:image", expectParseErr: true},
		{server: "lxd", image: "ubuntu:24.04", expectParse: true, expectImageSource: simplestreamsImage("https://cloud-images.ubuntu.com/releases/", "24.04")},
		{server: "lxd", image: "debian:12", expectParse: true, expectImageSource: simplestreamsImage("https://images.lxd.canonical.com", "debian/12/cloud")},
		{server: "lxd", image: "images:almalinux/9/cloud", expectParse: true, expectImageSource: simplestreamsImage("https://images.lxd.canonical.com", "almalinux/9/cloud")},
		{server: "lxd", image: "capi:kubeadm/v1.33.0", expectParse: true, expectImageSource: simplestreamsImage("https://images.capn.open-cloud.xyz", "kubeadm/v1.33.0")},
		{server: "lxd", image: "capi-stg:kubeadm/v1.33.0", expectParse: true, expectImageSource: simplestreamsImage("https://images-stg.capn.open-cloud.xyz", "kubeadm/v1.33.0")},
		{server: "lxd", image: "kind:v1.33.0", expectParse: true, expectImageSource: ociImage("https://docker.io", "kindest/node:v1.33.0")},
		// verify prefixes for unknown
		{server: "unknown", image: "image-name"},
		{server: "unknown", image: "ubuntu:24.04", expectParse: true, expectErr: true},
	} {
		t.Run(fmt.Sprintf("%s/%s", tc.server, tc.image), func(t *testing.T) {
			g := NewWithT(t)

			family, parsed, err := lxc.ParseImage(tc.image)
			if tc.expectParseErr {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).ToNot(HaveOccurred())
			if !tc.expectParse {
				g.Expect(parsed).To(BeFalse())
				return
			}

			image, err := family.For(tc.server)
			if tc.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(image.ImageSource()).To(Equal(tc.expectImageSource))
			}
		})
	}
}
