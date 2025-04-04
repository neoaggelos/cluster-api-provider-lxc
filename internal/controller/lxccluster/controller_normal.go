package lxccluster

import (
	"context"
	"fmt"

	"github.com/lxc/incus/v6/shared/api"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha2"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/incus"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/profile"
)

func (r *LXCClusterReconciler) reconcileNormal(ctx context.Context, cluster *clusterv1.Cluster, lxcCluster *infrav1.LXCCluster, lxcClient *incus.Client) error {
	// Create the default kubeadm profile for LXC containers
	profileName := lxcCluster.GetProfileName()
	if lxcCluster.Spec.SkipDefaultKubeadmProfile {
		conditions.MarkFalse(lxcCluster, infrav1.KubeadmProfileAvailableCondition, infrav1.KubeadmProfileDisabledReason, clusterv1.ConditionSeverityInfo, "Will not create default kubeadm profile %s", profileName)
	} else {
		// TODO: Canonical LXD still requires disabling apparmor because of:
		// [  830.196185] audit: type=1400 audit(1743727976.113:471): apparmor="DENIED" operation="pivotroot" class="mount" namespace="root//lxd-quick-start-unprivileged-82pwrt-vfvqs-n7cvw_<var-snap-lxd-common-lxd>" profile="runc" name="/run/containerd/io.containerd.runtime.v2.task/k8s.io/088164824e240ac135a312bedddf0503fc8a87d84ba45cceed3bcb9c8db510c8/rootfs/" pid=22340 comm="runc:[2:INIT]" srcname="/run/containerd/io.containerd.runtime.v2.task/k8s.io/088164824e240ac135a312bedddf0503fc8a87d84ba45cceed3bcb9c8db510c8/rootfs/"
		// containerd logs:
		// Apr 03 22:45:25 quick-start-unprivileged-8n4ykt-x4btr-5nrbl containerd[1421]: time="2025-04-03T22:45:25.763278581Z" level=error msg="RunPodSandbox for &PodSandboxMetadata{Name:kube-apiserver-quick-start-unprivileged-8n4ykt-x4btr-5nrbl,Uid:8cacae664d987ae932658b73b3c1fb47,Namespace:kube-system,Attempt:0,} failed, error" error="failed to create containerd task: failed to create shim task: OCI runtime create failed: runc create failed: unable to start container process: error during container init: error jailing process inside rootfs: pivot_root .: permission denied: unknown"
		// Apr 03 22:45:25 quick-start-unprivileged-8n4ykt-x4btr-5nrbl containerd[1421]: time="2025-04-03T22:45:25.787896658Z" level=info msg="shim disconnected" id=e868417d40bd232530ad8ade18304ec95b626b5a0dc8ea0abbcbcd1e08f3e05c namespace=k8s.io
		// Apr 03 22:45:25 quick-start-unprivileged-8n4ykt-x4btr-5nrbl containerd[1421]: time="2025-04-03T22:45:25.787926173Z" level=warning msg="cleaning up after shim disconnected" id=e868417d40bd232530ad8ade18304ec95b626b5a0dc8ea0abbcbcd1e08f3e05c namespace=k8s.io
		// Apr 03 22:45:25 quick-start-unprivileged-8n4ykt-x4btr-5nrbl containerd[1421]: time="2025-04-03T22:45:25.787934448Z" level=info msg="cleaning up dead shim" namespace=k8s.io
		// Apr 03 22:45:25 quick-start-unprivileged-8n4ykt-x4btr-5nrbl containerd[1421]: time="2025-04-03T22:45:25.797482148Z" level=warning msg="cleanup warnings time=\"2025-04-03T22:45:25Z\" level=warning msg=\"failed to read init pid file\" error=\"open /run/containerd/io.containerd.runtime.v2.task/k8s.io/e868417d40bd232530ad8ade18304ec95b626b5a0dc8ea0abbcbcd1e08f3e05c/init.pid: no such file or directory\" runtime=io.containerd.runc.v2\n" namespace=k8s.io
		// Apr 03 22:45:25 quick-start-unprivileged-8n4ykt-x4btr-5nrbl containerd[1421]: time="2025-04-03T22:45:25.798012123Z" level=error msg="copy shim log" error="read /proc/self/fd/21: file already closed" namespace=k8s.io
		// Apr 03 22:45:25 quick-start-unprivileged-8n4ykt-x4btr-5nrbl containerd[1421]: time="2025-04-03T22:45:25.802813105Z" level=error msg="RunPodSandbox for &PodSandboxMetadata{Name:kube-scheduler-quick-start-unprivileged-8n4ykt-x4btr-5nrbl,Uid:5a45eabb9a6c7f0b23e4aefb5397703f,Namespace:kube-system,Attempt:0,} failed, error" error="failed to create containerd task: failed to create shim task: OCI runtime create failed: runc create failed: unable to start container process: error during container init: error jailing process inside rootfs: pivot_root .: permission denied: unknown"
		if server, _, err := lxcClient.Client.GetServer(); err != nil {
			log.FromContext(ctx).Error(err, "Warning: Failed to check server information")
		} else if server.Environment.Server == "lxd" && lxcCluster.Spec.Unprivileged {
			conditions.MarkFalse(lxcCluster, infrav1.KubeadmProfileAvailableCondition, infrav1.KubeadmProfileCreationAbortedReason, clusterv1.ConditionSeverityError, "Unprivileged containers are currently only supported for Incus. Please use privileged containers, or specify instance type 'virtual-machine' for cluster nodes")
			return nil
		}

		ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("profileName", profileName, "privileged", !lxcCluster.Spec.Unprivileged))
		log.FromContext(ctx).Info("Creating default kubeadm profile")
		if err := lxcClient.InitProfile(ctx, api.ProfilesPost{Name: profileName, ProfilePut: profile.DefaultKubeadm(!lxcCluster.Spec.Unprivileged)}); err != nil {
			err = fmt.Errorf("failed to create default kubeadm profile %q: %w", profileName, err)
			log.FromContext(ctx).Error(err, "Failed to create default kubeadm profile")

			if incus.IsTerminalError(err) {
				conditions.MarkFalse(lxcCluster, infrav1.KubeadmProfileAvailableCondition, infrav1.KubeadmProfileCreationAbortedReason, clusterv1.ConditionSeverityError, "The default kubeadm LXC profile could not be created, most likely because of a permissions issue. Either enable privileged containers on the project, or specify .spec.skipDefaultKubeadmProfile=true on the LXCCluster object. The error was: %s", err)
				return nil
			}

			conditions.MarkFalse(lxcCluster, infrav1.KubeadmProfileAvailableCondition, infrav1.KubeadmProfileCreationFailedReason, clusterv1.ConditionSeverityWarning, "%s", err)
			return err
		}

		conditions.MarkTrue(lxcCluster, infrav1.KubeadmProfileAvailableCondition)
	}

	// Create the container hosting the load balancer.
	log.FromContext(ctx).Info("Creating load balancer")
	lbIPs, err := lxcClient.LoadBalancerManagerForCluster(cluster, lxcCluster).Create(ctx)
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to provision load balancer")
		if incus.IsTerminalError(err) {
			conditions.MarkFalse(lxcCluster, infrav1.LoadBalancerAvailableCondition, infrav1.LoadBalancerProvisioningAbortedReason, clusterv1.ConditionSeverityError, "The cluster load balancer could not be provisioned. The error was: %s", err)
			return nil
		}
		conditions.MarkFalse(lxcCluster, infrav1.LoadBalancerAvailableCondition, infrav1.LoadBalancerProvisioningFailedReason, clusterv1.ConditionSeverityWarning, "%s", err)
		return err
	}

	// Surface the control plane endpoint
	if lxcCluster.Spec.ControlPlaneEndpoint.Host == "" {
		// TODO(neoaggelos): care about IPv4 vs IPv6
		lxcCluster.Spec.ControlPlaneEndpoint.Host = lbIPs[0]
	}
	if lxcCluster.Spec.ControlPlaneEndpoint.Port == 0 {
		lxcCluster.Spec.ControlPlaneEndpoint.Port = 6443
	}

	// Mark the lxcCluster ready
	lxcCluster.Status.Ready = true
	conditions.MarkTrue(lxcCluster, infrav1.LoadBalancerAvailableCondition)

	return nil
}
