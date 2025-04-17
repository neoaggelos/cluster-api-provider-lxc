package static

import _ "embed"

var (
	//go:embed embed/kube-vip.tpl
	kubeVIPTemplate string

	//go:embed embed/kube-vip-hosts.tpl
	kubeVIPHostsTemplate string

	//go:embed embed/kube-flannel.tpl
	kubeFlannelTemplate string

	//go:embed embed/kube-proxy-config.tpl
	kubeProxyConfigTemplate string
)

func KubeVIPTemplate() string {
	return kubeVIPTemplate
}

func KubeVIPHostsTemplate() string {
	return kubeVIPHostsTemplate
}

func KubeFlannelTemplate() string {
	return kubeFlannelTemplate
}

func KubeProxyConfigTemplate() string {
	return kubeProxyConfigTemplate
}
