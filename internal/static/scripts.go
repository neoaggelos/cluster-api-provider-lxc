package static

import _ "embed"

var (
	//go:embed embed/install-kubeadm.sh
	installKubeadmScript string

	//go:embed embed/install-haproxy.sh
	installHaproxyScript string

	//go:embed embed/generate-manifest.sh
	generateManifestScript string

	//go:embed embed/cleanup-instance.sh
	cleanupInstanceScript string
)

func InstallKubeadmScript() string {
	return installKubeadmScript
}

func InstallHaproxyScript() string {
	return installHaproxyScript
}

func GenerateManifestScript() string {
	return generateManifestScript
}

func CleanupInstanceScript() string {
	return cleanupInstanceScript
}
