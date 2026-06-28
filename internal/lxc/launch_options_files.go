package lxc

import (
	"bytes"
	"fmt"

	incus "github.com/lxc/incus/v7/client"
)

type instanceFileCreator interface {
	path() string
	args() incus.InstanceFileArgs
	action() string
}

// createFile is an instanceFileCreator that creates a regular file.
type createFile struct {
	Path     string
	Contents string
}

func (f *createFile) path() string   { return f.Path }
func (f *createFile) action() string { return fmt.Sprintf("CreateInstanceFile(%q)", f.Path) }
func (f *createFile) args() incus.InstanceFileArgs {
	return incus.InstanceFileArgs{Content: bytes.NewReader([]byte(f.Contents)), Mode: 0644}
}

// createDirectory is an instanceFileCreator that creates a directory.
type createDirectory struct {
	Path string
}

func (f *createDirectory) path() string   { return f.Path }
func (f *createDirectory) action() string { return fmt.Sprintf("CreateDirectory(%q)", f.Path) }
func (f *createDirectory) args() incus.InstanceFileArgs {
	return incus.InstanceFileArgs{Type: "directory", Mode: 0755}
}

// createSymlink is an instanceFileCreator that creates a regular file.
type createSymlink struct {
	Path   string
	Target string
}

func (f *createSymlink) path() string { return f.Path }
func (f *createSymlink) action() string {
	return fmt.Sprintf("CreateSymlink(%q => %q)", f.Path, f.Target)
}
func (f *createSymlink) args() incus.InstanceFileArgs {
	return incus.InstanceFileArgs{Type: "symlink", Content: bytes.NewReader([]byte(f.Target)), Mode: 0644}
}

// appendTextToFile is an instanceFileCreator that appends text to an existing regular file.
type appendTextToFile struct {
	Path     string
	Contents string
}

func (f *appendTextToFile) path() string { return f.Path }
func (f *appendTextToFile) action() string {
	return fmt.Sprintf("AppendToFile(%q)", f.Path)
}
func (f *appendTextToFile) args() incus.InstanceFileArgs {
	return incus.InstanceFileArgs{Type: "file", Content: bytes.NewReader([]byte(f.Contents)), WriteMode: "append"}
}
