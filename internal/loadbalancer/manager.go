package loadbalancer

import (
	"context"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// Manager can be used to interact with the cluster load balancer.
type Manager interface {
	// Create provisions the load balancer instance.
	// Implementations can indicate non-retriable failures (e.g. because of Incus not having the required extensions).
	// Callers must check these with utils.IsTerminalError() and treat them as terminal failures.
	Create(context.Context) ([]string, error)
	// Delete cleans up any load balancer resources.
	Delete(context.Context) error
	// Reconfigure updates the load balancer configuration based on the currently running control plane instances.
	Reconfigure(context.Context) error
	// ControlPlaneInstanceTemplates is a map of files that will be injected as templates to control plane instances.
	ControlPlaneInstanceTemplates(controlPlaneInitialized bool) (map[string]string, error)
	// Inspect returns a map[string]string of the current state of the load balancer infrastructure.
	// It is mainly used by the E2E tests.
	Inspect(context.Context) map[string]string
}

// ManagerForCluster returns the proper Manager based on the lxcCluster spec.
func ManagerForCluster(cluster *clusterv1.Cluster, lxcCluster *infrav1.LXCCluster, lxcClient *lxc.Client) Manager {
	switch {
	case lxcCluster.Spec.LoadBalancer.LXC != nil:
		return &managerLXC{
			lxcClient:        lxcClient,
			clusterName:      cluster.Name,
			clusterNamespace: cluster.Namespace,

			name:                        lxcCluster.GetLoadBalancerInstanceName(),
			spec:                        lxcCluster.Spec.LoadBalancer.LXC.InstanceSpec,
			customHAProxyConfigTemplate: lxcCluster.Spec.LoadBalancer.LXC.CustomHAProxyConfigTemplate,
		}
	case lxcCluster.Spec.LoadBalancer.OCI != nil:
		return &managerOCI{
			lxcClient:        lxcClient,
			clusterName:      cluster.Name,
			clusterNamespace: cluster.Namespace,

			name:                        lxcCluster.GetLoadBalancerInstanceName(),
			spec:                        lxcCluster.Spec.LoadBalancer.OCI.InstanceSpec,
			customHAProxyConfigTemplate: lxcCluster.Spec.LoadBalancer.OCI.CustomHAProxyConfigTemplate,
		}
	case lxcCluster.Spec.LoadBalancer.OVN != nil:
		return &managerOVN{
			lxcClient:        lxcClient,
			clusterName:      cluster.Name,
			clusterNamespace: cluster.Namespace,

			networkName:   lxcCluster.Spec.LoadBalancer.OVN.NetworkName,
			listenAddress: lxcCluster.Spec.ControlPlaneEndpoint.Host,
		}
	case lxcCluster.Spec.LoadBalancer.External != nil:
		return &managerExternal{
			lxcClient:        lxcClient,
			clusterName:      cluster.Name,
			clusterNamespace: cluster.Namespace,

			address: lxcCluster.Spec.ControlPlaneEndpoint.Host,
		}
	case lxcCluster.Spec.LoadBalancer.KubeVIP != nil:
		return &managerKubeVIP{
			lxcClient:        lxcClient,
			clusterName:      cluster.Name,
			clusterNamespace: cluster.Namespace,

			address: lxcCluster.Spec.ControlPlaneEndpoint.Host,

			interfaceName:  lxcCluster.Spec.LoadBalancer.KubeVIP.Interface,
			image:          lxcCluster.Spec.LoadBalancer.KubeVIP.Image,
			kubeconfigPath: lxcCluster.Spec.LoadBalancer.KubeVIP.KubeconfigPath,
			manifestPath:   lxcCluster.Spec.LoadBalancer.KubeVIP.ManifestPath,
		}
	default:
		// TODO: handle this more gracefully.
		// If only Go had enums.
		return &managerExternal{
			lxcClient:        lxcClient,
			clusterName:      cluster.Name,
			clusterNamespace: cluster.Namespace,

			address: lxcCluster.Spec.ControlPlaneEndpoint.Host,
		}
	}
}
