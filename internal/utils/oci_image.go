package utils

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

type OCIImage struct {
	raw string
	ref name.Reference
}

// ParseOCIImage parses an OCI image name.
func ParseOCIImage(image string) (*OCIImage, error) {
	ref, err := name.ParseReference(image, name.WeakValidation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	return &OCIImage{
		raw: image,
		ref: ref,
	}, nil
}

// Server returns the HTTPS server endpoint of the image registry.
func (i *OCIImage) Server() string {
	registry := i.ref.Context().RegistryStr()
	if registry == "index.docker.io" {
		registry = "docker.io"
	}
	return fmt.Sprintf("%s://%s", i.ref.Context().Scheme(), registry)
}

// Alias returns the alias to use in Incus for pulling the image.
func (i *OCIImage) Alias() string {
	alias, _ := strings.CutPrefix(i.ref.Name(), i.ref.Context().RegistryStr()+"/")
	return alias
}

// Tag returns the image tag, if specified, or an empty string.
func (i *OCIImage) Tag() string {
	_, tag, _ := strings.Cut(i.raw, ":")
	tag, _, _ = strings.Cut(tag, "@")

	return tag
}

// Digest returns the image digest, if specified, or an empty string.
func (i *OCIImage) Digest() string {
	_, digest, _ := strings.Cut(i.raw, "@")
	return digest
}
