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

	//go:embed embed/validate-kubeadm-image.sh
	validateKubeadmImageScript string

	//go:embed embed/kind-cloud-init.py
	kindCloudInitScript string

	//go:embed embed/meta-data
	cloudInitMetaDataTemplate string

	//go:embed embed/user-data
	cloudInitUserDataTemplate string
)

func InstallKubeadmScript() string {
	return installKubeadmScript
}

func ValidateKubeadmImageScript() string {
	return validateKubeadmImageScript
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

func KindCloudInitScript() string {
	return kindCloudInitScript
}

func CloudInitMetaDataTemplate() string {
	return cloudInitMetaDataTemplate
}

func CloudInitUserDataTemplate() string {
	return cloudInitUserDataTemplate
}
