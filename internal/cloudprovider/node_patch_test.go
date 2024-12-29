package cloudprovider_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1 "k8s.io/api/core/v1"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/cloudprovider"
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
}
