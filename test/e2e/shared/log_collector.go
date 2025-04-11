//go:build e2e

package shared

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	expv1 "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/incus"
)

type IncusLogCollector struct {
	E2EContext *E2EContext
}

// CollectMachineLog gets logs for the LXC resources related to the given machine.
func (o IncusLogCollector) CollectMachineLog(ctx context.Context, managementClusterClient client.Client, m *clusterv1.Machine, outputPath string) error {
	Logf("Collecting logs for machine %q and storing them in %q", m.ObjectMeta.Name, outputPath)

	if err := os.MkdirAll(outputPath, 0o750); err != nil {
		return fmt.Errorf("couldn't create directory %q for logs: %s", outputPath, err)
	}

	lxcMachineName := types.NamespacedName{Name: m.Spec.InfrastructureRef.Name, Namespace: m.Spec.InfrastructureRef.Namespace}
	lxcMachine := &infrav1.LXCMachine{}
	if err := managementClusterClient.Get(ctx, types.NamespacedName{Name: m.Spec.InfrastructureRef.Name, Namespace: m.Spec.InfrastructureRef.Namespace}, lxcMachine); err != nil {
		return fmt.Errorf("failed to get LXCMachine %q for Machine: %w", lxcMachineName, err)
	}

	instanceName := lxcMachine.GetInstanceName()
	client, err := incus.New(ctx, o.E2EContext.Settings.LXCClientOptions)
	if err != nil {
		return fmt.Errorf("failed to create infrastructure client: %w", err)
	}

	// instance state and config
	{
		state, _, err := client.Client.GetInstanceFull(instanceName)
		if err != nil {
			return fmt.Errorf("failed to GetInstanceFull: %w", err)
		}

		if b, err := yaml.Marshal(state); err != nil {
			return fmt.Errorf("failed to marshal instance state: %w", err)
		} else if err := os.WriteFile(filepath.Join(outputPath, "instance.yaml"), b, 0o600); err != nil {
			return fmt.Errorf("failed to write instance.yaml: %w", err)
		}
	}

	type logitem struct {
		name    string
		command []string
	}

	items := []logitem{
		// version
		{name: "kubelet.version", command: []string{"kubelet", "--version"}},
		{name: "containerd.version", command: []string{"containerd", "--version"}},
		// system stuff
		{name: "system-df-h.log", command: []string{"df", "-h"}},
		{name: "system-uname-a.log", command: []string{"uname", "-a"}},
		{name: "system-ps-fea.log", command: []string{"ps", "-fea"}},
		{name: "system-ip-a.log", command: []string{"ip", "a"}},
		{name: "system-ip-r.log", command: []string{"ip", "r"}},
		{name: "system-iptables-save.log", command: []string{"iptables-save"}},
		{name: "system-lsmod.log", command: []string{"lsmod"}},
		{name: "system-sysctl-a.log", command: []string{"sysctl", "-a"}},
		{name: "system-free-h.log", command: []string{"free", "-h"}},
		// logs
		{name: "cloud-final.log", command: []string{"journalctl", "--no-pager", "-u", "cloud-final"}},
		{name: "kubelet.log", command: []string{"journalctl", "--no-pager", "-u", "kubelet.service"}},
		{name: "containerd.log", command: []string{"journalctl", "--no-pager", "-u", "containerd.service"}},
		{name: "var-log-pods.tar.gz", command: []string{"bash", "-c", "tar cv /var/log/pods | gzip"}},
		// kubernetes configuration
		{name: "etc-containerd.tar.gz", command: []string{"bash", "-c", "tar cv /etc/containerd | gzip"}},
		{name: "etc-kubernetes.tar.gz", command: []string{"bash", "-c", "tar cv /etc/kubernetes | gzip"}},
		{name: "run-kubeadm.tar.gz", command: []string{"bash", "-c", "tar cv /run/kubeadm | gzip"}},
		// container runtime
		{name: "crictl-info.json", command: []string{"crictl", "info"}},
		{name: "crictl-ps-a.log", command: []string{"crictl", "ps", "-a"}},
		{name: "crictl-pods.log", command: []string{"crictl", "pods"}},
	}

	// kernel logs (for virtual machines only)
	if lxcMachine.Spec.InstanceType == "virtual-machine" {
		items = append(items, logitem{name: "kern.log", command: []string{"journalctl", "--no-pager", "--output=short-precise", "-k"}})
	}

	var errs []error
	for _, item := range items {
		commandCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		var stdout, stderr bytes.Buffer
		if err := client.RunCommand(commandCtx, instanceName, item.command, &stdout, &stderr); err != nil {
			errs = append(errs, fmt.Errorf("failed to run command %v: %w", item.command, err))
		}
		if v := stdout.Bytes(); len(v) > 0 {
			if err := os.WriteFile(filepath.Join(outputPath, item.name), v, 0o600); err != nil {
				errs = append(errs, fmt.Errorf("failed to write stdout of command %v: %w", item.command, err))
			}
		}
		if v := stderr.Bytes(); len(v) > 0 {
			if err := os.WriteFile(filepath.Join(outputPath, item.name+".stderr"), v, 0o600); err != nil {
				errs = append(errs, fmt.Errorf("failed to write stderr of command %v: %w", item.command, err))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to collect logs from %s: %w", instanceName, errors.Join(errs...))
	}
	return nil
}

// CollectMachinePoolLog is not yet implemented for the LXC provider.
func (o IncusLogCollector) CollectMachinePoolLog(_ context.Context, _ client.Client, _ *expv1.MachinePool, _ string) error {
	return fmt.Errorf("not implemented")
}

// CollectInfrastructureLogs is not yet implemented for the LXC provider.
func (o IncusLogCollector) CollectInfrastructureLogs(ctx context.Context, managementClusterClient client.Client, cluster *clusterv1.Cluster, outputPath string) error {
	clusterName := types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}
	Logf("Collecting logs for cluster %q and storing them in %q", clusterName, outputPath)

	if err := os.MkdirAll(outputPath, 0o750); err != nil {
		return fmt.Errorf("couldn't create directory %q for logs: %s", outputPath, err)
	}

	lxcClusterName := types.NamespacedName{Name: cluster.Spec.InfrastructureRef.Name, Namespace: cluster.Spec.InfrastructureRef.Namespace}

	lxcCluster := &infrav1.LXCCluster{}
	if err := managementClusterClient.Get(ctx, lxcClusterName, lxcCluster); err != nil {
		return fmt.Errorf("failed to get LXCCluster %q for Cluster %q: %w", lxcClusterName, clusterName, err)
	}

	client, err := incus.New(ctx, o.E2EContext.Settings.LXCClientOptions)
	if err != nil {
		return fmt.Errorf("failed to create infrastructure client: %w", err)
	}

	var errs []error
	for k, v := range client.LoadBalancerManagerForCluster(cluster, lxcCluster).Inspect(ctx) {
		if err := os.WriteFile(filepath.Join(outputPath, k), []byte(v), 0o600); err != nil {
			errs = append(errs, fmt.Errorf("failed to write inspection file %v: %w", k, err))
		}
	}

	if errs != nil {
		return errors.Join(errs...)
	}
	return nil
}
