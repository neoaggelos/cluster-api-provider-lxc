package lxc

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/crane"
	incus "github.com/lxc/incus/v7/client"
	"github.com/lxc/incus/v7/shared/api"

	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

// ImageFamily is a well-known image for multiple servers (e.g. Incus, Canonical LXD).
type ImageFamily interface {
	For(serverName string) (Image, error)
}

// Image is an image source for instances.
type Image struct {
	Fingerprint string `json:"fingerprint,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	Server      string `json:"server,omitempty"`
	Alias       string `json:"alias,omitempty"`
}

type imageFamily struct {
	Description string           // description of image, e.g. "ubuntu"
	Sources     map[string]Image // server -> source
}

// For implements ImageFamily.
func (f *imageFamily) For(serverName string) (Image, error) {
	if img, ok := f.Sources[serverName]; !ok {
		return Image{}, utils.TerminalError(fmt.Errorf("no source for %q images for server %q", f.Description, serverName))
	} else {
		return img, nil
	}
}

// For implements ImageFamily.
func (i Image) For(serverName string) (Image, error) {
	return i, nil
}

func (i Image) ImageSource() api.ImageSource {
	return api.ImageSource{
		Protocol: i.Protocol,
		Server:   i.Server,
		Alias:    i.Alias,
	}
}

func (i Image) InstanceSource() api.InstanceSource {
	return api.InstanceSource{
		Type:        "image",
		Protocol:    i.Protocol,
		Server:      i.Server,
		Alias:       i.Alias,
		Fingerprint: i.Fingerprint,
	}
}

// Check runs a local check that the image is available on the upstream server.
//
// For simplestreams images, connects to the simplestreams server and validates image alias exists.
// For OCI images, validates that the HEAD request succeeds.
func (i *Image) Check(instanceType api.InstanceType) error {
	switch i.Protocol {
	case Simplestreams:
		if client, err := incus.ConnectSimpleStreams(i.Server, &incus.ConnectionArgs{HTTPClient: &http.Client{Timeout: 10 * time.Second}}); err != nil {
			return fmt.Errorf("failed to connect to simplestreams server %q: %w", i.Server, err)
		} else if _, _, err := client.GetImageAliasType(string(instanceType), i.Alias); err != nil {
			return utils.TerminalError(fmt.Errorf("no image with alias %q found on the simplestreams server %q: %w", i.Alias, i.Server, err))
		}
	case OCI:
		var opts []crane.Option
		var imageRef string
		if server, ok := strings.CutPrefix(i.Server, "https://"); ok {
			imageRef = fmt.Sprintf("%s/%s", server, i.Alias)
		} else if server, ok := strings.CutPrefix(i.Server, "http://"); ok {
			imageRef = fmt.Sprintf("%s/%s", server, i.Alias)
			opts = append(opts, crane.Insecure)
		} else {
			return utils.TerminalError(fmt.Errorf("server %q is not an HTTP or HTTPS server", i.Server))
		}

		if _, err := crane.Head(imageRef, opts...); err != nil {
			// example errors:
			// HEAD https://index.docker.io/v2/kindest/node/manifests/v1.34.0-not-exist: unexpected status code 404 Not Found (HEAD responses have no body, use GET for details)
			// HEAD https://index.docker.io/v2/kindest/node13131/manifests/v1.33.0: unexpected status code 401 Unauthorized (HEAD responses have no body, use GET for details)
			// HEAD http://w00:5050/v2/kindest/node13131/manifests/v1.33.0: unexpected status code 404 Not Found (HEAD responses have no body, use GET for details)
			err = fmt.Errorf("HEAD request for image %q failed: %w", imageRef, err)
			if !strings.Contains(err.Error(), "unexpected status code 4") {
				err = utils.TerminalError(fmt.Errorf("fatal error: %w", err))
			}
			return err
		}
	default:
		return fmt.Errorf("check not supported for protocol %q", i.Protocol)
	}
	return nil
}
