package lxc

import (
	"maps"

	"github.com/lxc/incus/v6/shared/api"
)

// LaunchOptions describe additional provisioning actions for machines.
type LaunchOptions struct {
	// instanceTemplates are "<file>"="<contents>" template files that will be created on the machine.
	// Supported by all instance types.
	instanceTemplates map[string]string
	// createFiles are files that will be created with CreateInstanceFile after creating the machine.
	// Not supported by virtual-machine instance types.
	createFiles []instanceFileCreator
	// replacements are a list of string replacements to perform on files on the machine.
	// The replacement is expected to be idempotent.
	// Not supported by virtual-machine instance types.
	replacements map[string]map[string]string
	// devices is instance device configuration.
	devices map[string]map[string]string
	// config is instance configuration.
	config map[string]string
	// profiles is instance profiles.
	profiles []string
	// image is the instance source.
	image ImageFamily
	// flavor is the instance flavor.
	flavor string
	// instanceType is the instance type.
	instanceType api.InstanceType
	// unixSocket bind mounts the admin unix socket into the instance at /run-unix.socket (potentially insecure).
	unixSocket bool
}

// WithInstanceTemplates appends instance templates.
func (o *LaunchOptions) WithInstanceTemplates(new map[string]string) *LaunchOptions {
	if o.instanceTemplates == nil {
		o.instanceTemplates = maps.Clone(new)
	} else {
		maps.Copy(o.instanceTemplates, new)
	}
	return o
}

// WithCreateFiles creates files on the instance.
func (o *LaunchOptions) WithCreateFiles(new map[string]string) *LaunchOptions {
	for path, content := range new {
		o.createFiles = append(o.createFiles, &createFile{
			Path:     path,
			Contents: content,
		})
	}
	return o
}

// WithAppendToFiles appends text to existing files on the instance.
func (o *LaunchOptions) WithAppendToFiles(new map[string]string) *LaunchOptions {
	for path, content := range new {
		o.createFiles = append(o.createFiles, &appendTextToFile{
			Path:     path,
			Contents: content,
		})
	}
	return o
}

// WithSymlinks creates symlinks on the instance.
func (o *LaunchOptions) WithSymlinks(new map[string]string) *LaunchOptions {
	for path, target := range new {
		o.createFiles = append(o.createFiles, &createSymlink{
			Path:   path,
			Target: target,
		})
	}
	return o
}

// WithDirectories creates directories on the instance.
func (o *LaunchOptions) WithDirectories(new ...string) *LaunchOptions {
	for _, path := range new {
		o.createFiles = append(o.createFiles, &createDirectory{Path: path})
	}
	return o
}

// WithReplacements appends instance file replacements.
func (o *LaunchOptions) WithReplacements(new map[string]map[string]string) *LaunchOptions {
	if o.replacements == nil {
		o.replacements = maps.Clone(new)
	} else {
		maps.Copy(o.replacements, new)
	}
	return o
}

// WithDevices adds instance devices.
func (o *LaunchOptions) WithDevices(new map[string]map[string]string) *LaunchOptions {
	if o.devices == nil {
		o.devices = maps.Clone(new)
	} else {
		maps.Copy(o.devices, new)
	}
	return o
}

// WithConfig adds instance config.
func (o *LaunchOptions) WithConfig(new map[string]string) *LaunchOptions {
	if o.config == nil {
		o.config = maps.Clone(new)
	} else {
		maps.Copy(o.config, new)
	}
	return o
}

// WithProfiles adds instance profiles.
func (o *LaunchOptions) WithProfiles(new []string) *LaunchOptions {
	o.profiles = append(o.profiles, new...)
	return o
}

// WithImage sets the instance image.
// WithImage is a no-op if an Image without an Alias or Fingerprint is passed.
func (o *LaunchOptions) WithImage(image ImageFamily) *LaunchOptions {
	if i, ok := image.(Image); !ok || len(i.Alias) > 0 || len(i.Fingerprint) > 0 {
		o.image = image
	}
	return o
}

// WithFlavor sets the instance flavor.
func (o *LaunchOptions) WithFlavor(v string) *LaunchOptions {
	o.flavor = v
	return o
}

// WithInstanceType sets the instance type (container or virtual-machine)
func (o *LaunchOptions) WithInstanceType(v api.InstanceType) *LaunchOptions {
	o.instanceType = v
	return o
}

// WithUnixSocket bind mounts the admin unix socket into the instance at /run-unix.socket (potentially insecure).
func (o *LaunchOptions) WithUnixSocket(v bool) *LaunchOptions {
	o.unixSocket = v
	return o
}
