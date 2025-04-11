/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lxcmachine_test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/controller/lxcmachine"

	. "github.com/onsi/gomega"
)

type cluster struct {
	cluster    *clusterv1.Cluster
	lxcCluster *infrav1.LXCCluster
}

type machine struct {
	machine    *clusterv1.Machine
	lxcMachine *infrav1.LXCMachine
}

func newCluster(name string, lxcName string) cluster {
	return cluster{
		cluster: &clusterv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: clusterv1.ClusterSpec{
				InfrastructureRef: &corev1.ObjectReference{
					APIVersion: infrav1.GroupVersion.String(),
					Kind:       "LXCCluster",
					Name:       lxcName,
				},
			},
		},
		lxcCluster: &infrav1.LXCCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: lxcName,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: clusterv1.GroupVersion.String(),
						Kind:       "Cluster",
						Name:       name,
					},
				},
			},
		},
	}
}

func newMachine(clusterName, name, lxcName string) machine {
	return machine{
		machine: &clusterv1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					clusterv1.ClusterNameLabel: clusterName,
				},
			},
			Spec: clusterv1.MachineSpec{
				InfrastructureRef: corev1.ObjectReference{
					APIVersion: infrav1.GroupVersion.String(),
					Kind:       "LXCMachine",
					Name:       lxcName,
				},
			},
		},
		lxcMachine: &infrav1.LXCMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name:       lxcName,
				Finalizers: []string{infrav1.MachineFinalizer},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: clusterv1.GroupVersion.String(),
						Kind:       "Machine",
						Name:       name,
					},
				},
			},
		},
	}
}

func newMachineWithoutLXC(clusterName, name string) machine {
	return machine{
		machine: &clusterv1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					clusterv1.ClusterNameLabel: clusterName,
				},
			},
		},
	}
}

var (
	cluster0 = newCluster("cluster-0", "lxc-cluster-0")
	cluster1 = newCluster("cluster-1", "lxc-cluster-1")
	cluster2 = newCluster("cluster-2", "lxc-cluster-2")

	machine0 = newMachine("cluster-0", "machine-0", "lxc-machine-0")
	machine1 = newMachine("cluster-0", "machine-1", "lxc-machine-1")
	machine2 = newMachineWithoutLXC("cluster-1", "machine-2")
	machine3 = newMachine("cluster-1", "machine-3", "lxc-machine-3")
	machine4 = newMachine("cluster-1", "machine-4", "lxc-machine-4")
	machine5 = newMachineWithoutLXC("cluster-1", "machine-5")
)

// TODO(neoaggelos): this should use ginkgo
func TestLXCMachineReconciler_LXCClusterToLXCMachines(t *testing.T) {
	g := NewWithT(t)

	objects := []client.Object{
		cluster0.cluster,
		cluster0.lxcCluster,
		cluster1.cluster,
		cluster1.lxcCluster,
		machine0.machine,
		machine0.lxcMachine,
		machine1.machine,
		machine1.lxcMachine,
		machine2.machine,
		// machine2.lxcMachine,  // NOTE(neoaggelos): omitted to test Machines without an InstructureRef
		machine3.machine,
		machine3.lxcMachine,
		machine4.machine,
		machine4.lxcMachine,
		machine5.machine,
		// machine5.lxcMachine,  // NOTE(neoaggelos): omitted to test Machines without an InstructureRef
	}
	c := fake.NewClientBuilder().WithObjects(objects...).Build()
	r := lxcmachine.LXCMachineReconciler{
		Client: c,
	}
	g.Expect(r.LXCClusterToLXCMachines(context.TODO(), cluster0.lxcCluster)).To(ConsistOf(
		HaveField("Name", "lxc-machine-0"),
		HaveField("Name", "lxc-machine-1"),
	))
	g.Expect(r.LXCClusterToLXCMachines(context.TODO(), cluster1.lxcCluster)).To(ConsistOf(
		HaveField("Name", "lxc-machine-3"),
		HaveField("Name", "lxc-machine-4"),
	))

	g.Expect(r.LXCClusterToLXCMachines(context.TODO(), cluster2.lxcCluster)).To(BeEmpty())
}

func init() {
	_ = clusterv1.AddToScheme(scheme.Scheme)
	_ = infrav1.AddToScheme(scheme.Scheme)
}
