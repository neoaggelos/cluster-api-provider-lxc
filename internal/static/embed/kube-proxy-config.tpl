---
kind: KubeProxyConfiguration
apiVersion: kubeproxy.config.k8s.io/v1alpha1
mode: iptables
conntrack:
  maxPerCore: 0
