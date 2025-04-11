package cloudprovider_test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/cloudprovider"

	. "github.com/onsi/gomega"
)

func TestPatchNode(t *testing.T) {
	t.Run("SetProviderID", func(t *testing.T) {
		remoteNode := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node0"},
		}
		lxcMachine := &infrav1.LXCMachine{
			ObjectMeta: metav1.ObjectMeta{Name: "node0", Namespace: "default"},
		}

		remoteClient := fake.NewFakeClient(remoteNode)
		g := NewWithT(t)
		g.Expect(cloudprovider.PatchNode(context.TODO(), remoteClient, lxcMachine)).NotTo(HaveOccurred())

		node := &corev1.Node{}
		g.Expect(remoteClient.Get(context.TODO(), client.ObjectKeyFromObject(remoteNode), node)).NotTo(HaveOccurred())
		g.Expect(node.Spec.ProviderID).To(Equal("lxc:///node0"))
	})

	t.Run("CloudProviderTaint", func(t *testing.T) {
		remoteNode := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node0"},
			Spec: corev1.NodeSpec{
				Taints: []corev1.Taint{
					{Key: "something", Effect: "NoSchedule"},
					{Key: "node.cloudprovider.kubernetes.io/uninitialized", Effect: "NoSchedule"},
				},
			},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeHostName, Address: "node0"},
					{Type: corev1.NodeInternalIP, Address: "10.0.0.20"},
				},
			},
		}
		lxcMachine := &infrav1.LXCMachine{
			ObjectMeta: metav1.ObjectMeta{Name: "node0", Namespace: "default"},
			Status: infrav1.LXCMachineStatus{
				Addresses: []clusterv1.MachineAddress{
					{Type: clusterv1.MachineHostName, Address: "node0"},
					{Type: clusterv1.MachineInternalIP, Address: "10.0.0.10"},
					{Type: clusterv1.MachineExternalIP, Address: "10.0.0.10"},
					{Type: clusterv1.MachineInternalIP, Address: "10.0.0.20"},
					{Type: clusterv1.MachineExternalIP, Address: "10.0.0.20"},
				},
			},
		}

		remoteClient := fake.NewFakeClient(remoteNode)
		g := NewWithT(t)
		g.Expect(cloudprovider.PatchNode(context.TODO(), remoteClient, lxcMachine)).NotTo(HaveOccurred())

		node := &corev1.Node{}
		g.Expect(remoteClient.Get(context.TODO(), client.ObjectKeyFromObject(remoteNode), node)).NotTo(HaveOccurred())
		g.Expect(node.Spec.Taints).To(ConsistOf(
			// cloud provider taint removed, other taints remain
			corev1.Taint{Key: "something", Effect: "NoSchedule"}),
		)
		g.Expect(node.Status.Addresses).To(ConsistOf(
			// original addresses
			corev1.NodeAddress{Type: corev1.NodeHostName, Address: "node0"},
			corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.20"},
			// missing InternalIP address appended
			corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.10"},
		))
	})
}
