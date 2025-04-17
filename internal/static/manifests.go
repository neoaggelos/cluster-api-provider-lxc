package static

import _ "embed"

var (
	//go:embed embed/kube-proxy-config.tpl
	kubeProxyConfigTemplate string
)

func KubeProxyConfigTemplate() string {
	return kubeProxyConfigTemplate
}
