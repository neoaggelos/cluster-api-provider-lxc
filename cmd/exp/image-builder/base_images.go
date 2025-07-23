package main

type baseImageInfo struct {
	fullName    string
	releaseName string
	variantName string
}

var wellKnownBaseImages = map[string]baseImageInfo{
	"ubuntu:20.04": {
		fullName:    "ubuntu focal",
		releaseName: "focal",
		variantName: "ubuntu",
	},
	"ubuntu:22.04": {
		fullName:    "ubuntu jammy",
		releaseName: "jammy",
		variantName: "ubuntu",
	},
	"ubuntu:24.04": {
		fullName:    "ubuntu noble",
		releaseName: "noble",
		variantName: "ubuntu",
	},
	"debian:12": {
		fullName:    "debian bookworm",
		releaseName: "bookworm",
		variantName: "debian",
	},
}
