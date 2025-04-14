package static

import _ "embed"

var (
	//go:embed embed/kube-flannel.tpl
	kubeFlannelTemplate string

	//go:embed embed/kube-proxy-config.tpl
	kubeProxyConfigTemplate string
)

func KubeFlannelTemplate() string {
	return kubeFlannelTemplate
}

func KubeProxyConfigTemplate() string {
	return kubeProxyConfigTemplate
}
