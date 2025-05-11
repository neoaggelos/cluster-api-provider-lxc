<h2 id="infrastructure.cluster.x-k8s.io/v1alpha2">infrastructure.cluster.x-k8s.io/v1alpha2</h2>
<p>
<p>package v1alpha2 contains API Schema definitions for the infrastructure v1alpha2 API group</p>
</p>
Resource Types:
<ul></ul>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCCluster">LXCCluster
</h3>
<p>
<p>LXCCluster is the Schema for the lxcclusters API.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterSpec">
LXCClusterSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>controlPlaneEndpoint</code><br/>
<em>
<a href="https://doc.crds.dev/github.com/kubernetes-sigs/cluster-api@v1.10.1">
sigs.k8s.io/cluster-api/api/v1beta1.APIEndpoint
</a>
</em>
</td>
<td>
<p>ControlPlaneEndpoint represents the endpoint to communicate with the control plane.</p>
</td>
</tr>
<tr>
<td>
<code>secretRef</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.SecretRef">
SecretRef
</a>
</em>
</td>
<td>
<p>SecretRef references a secret with credentials to access the LXC (e.g. Incus, LXD) server.</p>
</td>
</tr>
<tr>
<td>
<code>loadBalancer</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterLoadBalancer">
LXCClusterLoadBalancer
</a>
</em>
</td>
<td>
<p>LoadBalancer is configuration for provisioning the load balancer of the cluster.</p>
</td>
</tr>
<tr>
<td>
<code>unprivileged</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Unprivileged will launch unprivileged LXC containers for the cluster machines.</p>
<p>Known limitations apply for unprivileged LXC containers (e.g. cannot use NFS volumes).</p>
</td>
</tr>
<tr>
<td>
<code>skipDefaultKubeadmProfile</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Skip creation of the default kubeadm profile &ldquo;cluster-api-$namespace-$name&rdquo;
for LXCClusters.</p>
<p>In this case, the cluster administrator is responsible to create the
profile manually and set the <code>.spec.template.spec.profiles</code> field of all
LXCMachineTemplate objects.</p>
<p>This is useful in cases where a restricted project is used, which does not
allow privileged containers.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterStatus">
LXCClusterStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterLoadBalancer">LXCClusterLoadBalancer
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterSpec">LXCClusterSpec</a>)
</p>
<p>
<p>LXCClusterLoadBalancer is configuration for provisioning the load balancer of the cluster.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>lxc</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCLoadBalancerInstance">
LXCLoadBalancerInstance
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LXC will spin up a plain Ubuntu instance with haproxy installed.</p>
<p>The controller will automatically update the list of backends on the haproxy configuration as control plane nodes are added or removed from the cluster.</p>
<p>No other configuration is required for &ldquo;lxc&rdquo; mode. The load balancer instance can be configured through the .instanceSpec field.</p>
<p>The load balancer container is a single point of failure to access the workload cluster control plane. Therefore, it should only be used for development or evaluation clusters.</p>
</td>
</tr>
<tr>
<td>
<code>oci</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCLoadBalancerInstance">
LXCLoadBalancerInstance
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>OCI will spin up an OCI instance running the kindest/haproxy image.</p>
<p>The controller will automatically update the list of backends on the haproxy configuration as control plane nodes are added or removed from the cluster.</p>
<p>No other configuration is required for &ldquo;oci&rdquo; mode. The load balancer instance can be configured through the .instanceSpec field.</p>
<p>The load balancer container is a single point of failure to access the workload cluster control plane. Therefore, it should only be used for development or evaluation clusters.</p>
<p>Requires server extensions: &ldquo;instance_oci&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>ovn</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCLoadBalancerOVN">
LXCLoadBalancerOVN
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>OVN will create a network load balancer.</p>
<p>The controller will automatically update the list of backends for the network load balancer as control plane nodes are added or removed from the cluster.</p>
<p>The cluster administrator is responsible to ensure that the OVN network is configured properly and that the LXCMachineTemplate objects have appropriate profiles to use the OVN network.</p>
<p>When using the &ldquo;ovn&rdquo; mode, the load balancer address must be set in <code>.spec.controlPlaneEndpoint.host</code> on the LXCCluster object.</p>
<p>Requires server extensions: &ldquo;network_load_balancer&rdquo;, &ldquo;network_load_balancer_health_checks&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>external</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCLoadBalancerExternal">
LXCLoadBalancerExternal
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>External will not create a load balancer. It must be used alongside something like kube-vip, otherwise the cluster will fail to provision.</p>
<p>When using the &ldquo;external&rdquo; mode, the load balancer address must be set in <code>.spec.controlPlaneEndpoint.host</code> on the LXCCluster object.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterSpec">LXCClusterSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCCluster">LXCCluster</a>, 
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterTemplateResource">LXCClusterTemplateResource</a>)
</p>
<p>
<p>LXCClusterSpec defines the desired state of LXCCluster.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>controlPlaneEndpoint</code><br/>
<em>
<a href="https://doc.crds.dev/github.com/kubernetes-sigs/cluster-api@v1.10.1">
sigs.k8s.io/cluster-api/api/v1beta1.APIEndpoint
</a>
</em>
</td>
<td>
<p>ControlPlaneEndpoint represents the endpoint to communicate with the control plane.</p>
</td>
</tr>
<tr>
<td>
<code>secretRef</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.SecretRef">
SecretRef
</a>
</em>
</td>
<td>
<p>SecretRef references a secret with credentials to access the LXC (e.g. Incus, LXD) server.</p>
</td>
</tr>
<tr>
<td>
<code>loadBalancer</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterLoadBalancer">
LXCClusterLoadBalancer
</a>
</em>
</td>
<td>
<p>LoadBalancer is configuration for provisioning the load balancer of the cluster.</p>
</td>
</tr>
<tr>
<td>
<code>unprivileged</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Unprivileged will launch unprivileged LXC containers for the cluster machines.</p>
<p>Known limitations apply for unprivileged LXC containers (e.g. cannot use NFS volumes).</p>
</td>
</tr>
<tr>
<td>
<code>skipDefaultKubeadmProfile</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Skip creation of the default kubeadm profile &ldquo;cluster-api-$namespace-$name&rdquo;
for LXCClusters.</p>
<p>In this case, the cluster administrator is responsible to create the
profile manually and set the <code>.spec.template.spec.profiles</code> field of all
LXCMachineTemplate objects.</p>
<p>This is useful in cases where a restricted project is used, which does not
allow privileged containers.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterStatus">LXCClusterStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCCluster">LXCCluster</a>)
</p>
<p>
<p>LXCClusterStatus defines the observed state of LXCCluster.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ready</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ready denotes that the LXC cluster (infrastructure) is ready.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="https://doc.crds.dev/github.com/kubernetes-sigs/cluster-api@v1.10.1">
sigs.k8s.io/cluster-api/api/v1beta1.Conditions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Conditions defines current service state of the LXCCluster.</p>
</td>
</tr>
<tr>
<td>
<code>v1beta2</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterV1Beta2Status">
LXCClusterV1Beta2Status
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>V1Beta2 groups all status fields that will be added in LXCCluster&rsquo;s status with the v1beta2 version.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterTemplate">LXCClusterTemplate
</h3>
<p>
<p>LXCClusterTemplate is the Schema for the lxcclustertemplates API.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterTemplateSpec">
LXCClusterTemplateSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>template</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterTemplateResource">
LXCClusterTemplateResource
</a>
</em>
</td>
<td>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterTemplateResource">LXCClusterTemplateResource
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterTemplateSpec">LXCClusterTemplateSpec</a>)
</p>
<p>
<p>LXCClusterTemplateResource describes the data needed to create a LXCCluster from a template.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://doc.crds.dev/github.com/kubernetes-sigs/cluster-api@v1.10.1">
sigs.k8s.io/cluster-api/api/v1beta1.ObjectMeta
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Standard object&rsquo;s metadata.
More info: <a href="https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata">https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata</a></p>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterSpec">
LXCClusterSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>controlPlaneEndpoint</code><br/>
<em>
<a href="https://doc.crds.dev/github.com/kubernetes-sigs/cluster-api@v1.10.1">
sigs.k8s.io/cluster-api/api/v1beta1.APIEndpoint
</a>
</em>
</td>
<td>
<p>ControlPlaneEndpoint represents the endpoint to communicate with the control plane.</p>
</td>
</tr>
<tr>
<td>
<code>secretRef</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.SecretRef">
SecretRef
</a>
</em>
</td>
<td>
<p>SecretRef references a secret with credentials to access the LXC (e.g. Incus, LXD) server.</p>
</td>
</tr>
<tr>
<td>
<code>loadBalancer</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterLoadBalancer">
LXCClusterLoadBalancer
</a>
</em>
</td>
<td>
<p>LoadBalancer is configuration for provisioning the load balancer of the cluster.</p>
</td>
</tr>
<tr>
<td>
<code>unprivileged</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Unprivileged will launch unprivileged LXC containers for the cluster machines.</p>
<p>Known limitations apply for unprivileged LXC containers (e.g. cannot use NFS volumes).</p>
</td>
</tr>
<tr>
<td>
<code>skipDefaultKubeadmProfile</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Skip creation of the default kubeadm profile &ldquo;cluster-api-$namespace-$name&rdquo;
for LXCClusters.</p>
<p>In this case, the cluster administrator is responsible to create the
profile manually and set the <code>.spec.template.spec.profiles</code> field of all
LXCMachineTemplate objects.</p>
<p>This is useful in cases where a restricted project is used, which does not
allow privileged containers.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterTemplateSpec">LXCClusterTemplateSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterTemplate">LXCClusterTemplate</a>)
</p>
<p>
<p>LXCClusterTemplateSpec defines the desired state of LXCClusterTemplate.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>template</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterTemplateResource">
LXCClusterTemplateResource
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterV1Beta2Status">LXCClusterV1Beta2Status
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterStatus">LXCClusterStatus</a>)
</p>
<p>
<p>LXCClusterV1Beta2Status groups all the fields that will be added or modified in LXCCluster with the V1Beta2 version.
See <a href="https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20240916-improve-status-in-CAPI-resources.md">https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20240916-improve-status-in-CAPI-resources.md</a> for more context.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Condition">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>conditions represents the observations of a LXCCluster&rsquo;s current state.
Known condition types are Ready, LoadBalancerAvailable, Deleting, Paused.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCLoadBalancerExternal">LXCLoadBalancerExternal
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterLoadBalancer">LXCClusterLoadBalancer</a>)
</p>
<p>
</p>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCLoadBalancerInstance">LXCLoadBalancerInstance
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterLoadBalancer">LXCClusterLoadBalancer</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>instanceSpec</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCLoadBalancerMachineSpec">
LXCLoadBalancerMachineSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>InstanceSpec can be used to adjust the load balancer instance configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCLoadBalancerMachineSpec">LXCLoadBalancerMachineSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCLoadBalancerInstance">LXCLoadBalancerInstance</a>)
</p>
<p>
<p>LXCLoadBalancerMachineSpec is configuration for the container that will host the cluster load balancer, when using the &ldquo;lxc&rdquo; or &ldquo;oci&rdquo; load balancer type.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>flavor</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flavor is configuration for the instance size (e.g. t3.micro, or c2-m4).</p>
<p>Examples:</p>
<ul>
<li><code>t3.micro</code> &ndash; match specs of an EC2 t3.micro instance</li>
<li><code>c2-m4</code> &ndash; 2 cores, 4 GB RAM</li>
</ul>
</td>
</tr>
<tr>
<td>
<code>profiles</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Profiles is a list of profiles to attach to the instance.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineImageSource">
LXCMachineImageSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image to use for provisioning the load balancer machine. If not set,
a default image based on the load balancer type will be used.</p>
<ul>
<li>&ldquo;oci&rdquo;: ghcr.io/neoaggelos/cluster-api-provider-lxc/haproxy:v0.0.1</li>
<li>&ldquo;lxc&rdquo;: haproxy from the default simplestreams server</li>
</ul>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCLoadBalancerOVN">LXCLoadBalancerOVN
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterLoadBalancer">LXCClusterLoadBalancer</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>networkName</code><br/>
<em>
string
</em>
</td>
<td>
<p>NetworkName is the name of the network to create the load balancer.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachine">LXCMachine
</h3>
<p>
<p>LXCMachine is the Schema for the lxcmachines API.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineSpec">
LXCMachineSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>providerID</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProviderID is the container name in ProviderID format (lxc:///<containername>).</p>
</td>
</tr>
<tr>
<td>
<code>instanceType</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>InstanceType is &ldquo;container&rdquo; or &ldquo;virtual-machine&rdquo;. Empty defaults to &ldquo;container&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>flavor</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flavor is configuration for the instance size (e.g. t3.micro, or c2-m4).</p>
<p>Examples:</p>
<ul>
<li><code>t3.micro</code> &ndash; match specs of an EC2 t3.micro instance</li>
<li><code>c2-m4</code> &ndash; 2 cores, 4 GB RAM</li>
</ul>
</td>
</tr>
<tr>
<td>
<code>profiles</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Profiles is a list of profiles to attach to the instance.</p>
</td>
</tr>
<tr>
<td>
<code>devices</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Devices allows overriding the configuration of the instance disk or network.</p>
<p>Device configuration must be formatted using the syntax &ldquo;<device>,<key>=<value>&rdquo;.</p>
<p>For example, to specify a different network for an instance, you can use:</p>
<pre><code class="language-yaml">  # override device &quot;eth0&quot;, to be of type &quot;nic&quot; and use network &quot;my-network&quot;
devices:
- eth0,type=nic,network=my-network
</code></pre>
</td>
</tr>
<tr>
<td>
<code>config</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Config allows overriding instance configuration keys.</p>
<p>Note that the provider will always set the following configuration keys:</p>
<ul>
<li>&ldquo;cloud-init.user-data&rdquo;: cloud-init config data</li>
<li>&ldquo;user.cluster-name&rdquo;: name of owning cluster</li>
<li>&ldquo;user.cluster-namespace&rdquo;: namespace of owning cluster</li>
<li>&ldquo;user.cluster-role&rdquo;: instance role (e.g. control-plane, worker)</li>
<li>&ldquo;user.machine-name&rdquo;: name of machine (should match instance hostname)</li>
</ul>
<p>See <a href="https://linuxcontainers.org/incus/docs/main/reference/instance_options/#instance-options">https://linuxcontainers.org/incus/docs/main/reference/instance_options/#instance-options</a>
for details.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineImageSource">
LXCMachineImageSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image to use for provisioning the machine. If not set, a kubeadm image
from the default upstream simplestreams source will be used, based on
the version of the machine.</p>
<p>Note that the default source does not support images for all Kubernetes
versions, refer to the documentation for more details on which versions
are supported and how to build a base image for any version.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineStatus">
LXCMachineStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineImageSource">LXCMachineImageSource
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCLoadBalancerMachineSpec">LXCLoadBalancerMachineSpec</a>, 
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineSpec">LXCMachineSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name is the image name or alias.</p>
<p>Note that Incus and Canonical LXD use incompatible image servers
for Ubuntu images. To address this issue, setting image name to
<code>ubuntu:VERSION</code> is a shortcut for:</p>
<ul>
<li>Incus: &ldquo;images:ubuntu/VERSION/cloud&rdquo; (from <a href="https://images.linuxcontainers.org">https://images.linuxcontainers.org</a>)</li>
<li>LXD: &ldquo;ubuntu:VERSION&rdquo; (from <a href="https://cloud-images.ubuntu.com/releases">https://cloud-images.ubuntu.com/releases</a>)</li>
</ul>
</td>
</tr>
<tr>
<td>
<code>fingerprint</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Fingerprint is the image fingerprint.</p>
</td>
</tr>
<tr>
<td>
<code>server</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Server is the remote server, e.g. &ldquo;<a href="https://images.linuxcontainers.org&quot;">https://images.linuxcontainers.org&rdquo;</a></p>
</td>
</tr>
<tr>
<td>
<code>protocol</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Protocol is the protocol to use for fetching the image, e.g. &ldquo;simplestreams&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineSpec">LXCMachineSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachine">LXCMachine</a>, 
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineTemplateResource">LXCMachineTemplateResource</a>)
</p>
<p>
<p>LXCMachineSpec defines the desired state of LXCMachine.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>providerID</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProviderID is the container name in ProviderID format (lxc:///<containername>).</p>
</td>
</tr>
<tr>
<td>
<code>instanceType</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>InstanceType is &ldquo;container&rdquo; or &ldquo;virtual-machine&rdquo;. Empty defaults to &ldquo;container&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>flavor</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flavor is configuration for the instance size (e.g. t3.micro, or c2-m4).</p>
<p>Examples:</p>
<ul>
<li><code>t3.micro</code> &ndash; match specs of an EC2 t3.micro instance</li>
<li><code>c2-m4</code> &ndash; 2 cores, 4 GB RAM</li>
</ul>
</td>
</tr>
<tr>
<td>
<code>profiles</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Profiles is a list of profiles to attach to the instance.</p>
</td>
</tr>
<tr>
<td>
<code>devices</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Devices allows overriding the configuration of the instance disk or network.</p>
<p>Device configuration must be formatted using the syntax &ldquo;<device>,<key>=<value>&rdquo;.</p>
<p>For example, to specify a different network for an instance, you can use:</p>
<pre><code class="language-yaml">  # override device &quot;eth0&quot;, to be of type &quot;nic&quot; and use network &quot;my-network&quot;
devices:
- eth0,type=nic,network=my-network
</code></pre>
</td>
</tr>
<tr>
<td>
<code>config</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Config allows overriding instance configuration keys.</p>
<p>Note that the provider will always set the following configuration keys:</p>
<ul>
<li>&ldquo;cloud-init.user-data&rdquo;: cloud-init config data</li>
<li>&ldquo;user.cluster-name&rdquo;: name of owning cluster</li>
<li>&ldquo;user.cluster-namespace&rdquo;: namespace of owning cluster</li>
<li>&ldquo;user.cluster-role&rdquo;: instance role (e.g. control-plane, worker)</li>
<li>&ldquo;user.machine-name&rdquo;: name of machine (should match instance hostname)</li>
</ul>
<p>See <a href="https://linuxcontainers.org/incus/docs/main/reference/instance_options/#instance-options">https://linuxcontainers.org/incus/docs/main/reference/instance_options/#instance-options</a>
for details.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineImageSource">
LXCMachineImageSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image to use for provisioning the machine. If not set, a kubeadm image
from the default upstream simplestreams source will be used, based on
the version of the machine.</p>
<p>Note that the default source does not support images for all Kubernetes
versions, refer to the documentation for more details on which versions
are supported and how to build a base image for any version.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineStatus">LXCMachineStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachine">LXCMachine</a>)
</p>
<p>
<p>LXCMachineStatus defines the observed state of LXCMachine.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ready</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ready denotes that the LXC machine is ready.</p>
</td>
</tr>
<tr>
<td>
<code>loadBalancerConfigured</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>LoadBalancerConfigured will be set to true once for each control plane node, after the load balancer instance is reconfigured.</p>
</td>
</tr>
<tr>
<td>
<code>addresses</code><br/>
<em>
<a href="https://doc.crds.dev/github.com/kubernetes-sigs/cluster-api@v1.10.1">
[]sigs.k8s.io/cluster-api/api/v1beta1.MachineAddress
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Addresses is the list of addresses of the LXC machine.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="https://doc.crds.dev/github.com/kubernetes-sigs/cluster-api@v1.10.1">
sigs.k8s.io/cluster-api/api/v1beta1.Conditions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Conditions defines current service state of the LXCMachine.</p>
</td>
</tr>
<tr>
<td>
<code>v1beta2</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineV1Beta2Status">
LXCMachineV1Beta2Status
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>V1Beta2 groups all status fields that will be added in LXCMachine&rsquo;s status with the v1beta2 version.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineTemplate">LXCMachineTemplate
</h3>
<p>
<p>LXCMachineTemplate is the Schema for the lxcmachinetemplates API.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineTemplateSpec">
LXCMachineTemplateSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>template</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineTemplateResource">
LXCMachineTemplateResource
</a>
</em>
</td>
<td>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineTemplateResource">LXCMachineTemplateResource
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineTemplateSpec">LXCMachineTemplateSpec</a>)
</p>
<p>
<p>LXCMachineTemplateResource describes the data needed to create a LXCMachine from a template.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://doc.crds.dev/github.com/kubernetes-sigs/cluster-api@v1.10.1">
sigs.k8s.io/cluster-api/api/v1beta1.ObjectMeta
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Standard object&rsquo;s metadata.
More info: <a href="https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata">https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata</a></p>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineSpec">
LXCMachineSpec
</a>
</em>
</td>
<td>
<p>Spec is the specification of the desired behavior of the machine.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>providerID</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProviderID is the container name in ProviderID format (lxc:///<containername>).</p>
</td>
</tr>
<tr>
<td>
<code>instanceType</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>InstanceType is &ldquo;container&rdquo; or &ldquo;virtual-machine&rdquo;. Empty defaults to &ldquo;container&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>flavor</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flavor is configuration for the instance size (e.g. t3.micro, or c2-m4).</p>
<p>Examples:</p>
<ul>
<li><code>t3.micro</code> &ndash; match specs of an EC2 t3.micro instance</li>
<li><code>c2-m4</code> &ndash; 2 cores, 4 GB RAM</li>
</ul>
</td>
</tr>
<tr>
<td>
<code>profiles</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Profiles is a list of profiles to attach to the instance.</p>
</td>
</tr>
<tr>
<td>
<code>devices</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Devices allows overriding the configuration of the instance disk or network.</p>
<p>Device configuration must be formatted using the syntax &ldquo;<device>,<key>=<value>&rdquo;.</p>
<p>For example, to specify a different network for an instance, you can use:</p>
<pre><code class="language-yaml">  # override device &quot;eth0&quot;, to be of type &quot;nic&quot; and use network &quot;my-network&quot;
devices:
- eth0,type=nic,network=my-network
</code></pre>
</td>
</tr>
<tr>
<td>
<code>config</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Config allows overriding instance configuration keys.</p>
<p>Note that the provider will always set the following configuration keys:</p>
<ul>
<li>&ldquo;cloud-init.user-data&rdquo;: cloud-init config data</li>
<li>&ldquo;user.cluster-name&rdquo;: name of owning cluster</li>
<li>&ldquo;user.cluster-namespace&rdquo;: namespace of owning cluster</li>
<li>&ldquo;user.cluster-role&rdquo;: instance role (e.g. control-plane, worker)</li>
<li>&ldquo;user.machine-name&rdquo;: name of machine (should match instance hostname)</li>
</ul>
<p>See <a href="https://linuxcontainers.org/incus/docs/main/reference/instance_options/#instance-options">https://linuxcontainers.org/incus/docs/main/reference/instance_options/#instance-options</a>
for details.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineImageSource">
LXCMachineImageSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image to use for provisioning the machine. If not set, a kubeadm image
from the default upstream simplestreams source will be used, based on
the version of the machine.</p>
<p>Note that the default source does not support images for all Kubernetes
versions, refer to the documentation for more details on which versions
are supported and how to build a base image for any version.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineTemplateSpec">LXCMachineTemplateSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineTemplate">LXCMachineTemplate</a>)
</p>
<p>
<p>LXCMachineTemplateSpec defines the desired state of LXCMachineTemplate.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>template</code><br/>
<em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineTemplateResource">
LXCMachineTemplateResource
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineV1Beta2Status">LXCMachineV1Beta2Status
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCMachineStatus">LXCMachineStatus</a>)
</p>
<p>
<p>LXCMachineV1Beta2Status groups all the fields that will be added or modified in LXCMachine with the V1Beta2 version.
See <a href="https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20240916-improve-status-in-CAPI-resources.md">https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20240916-improve-status-in-CAPI-resources.md</a> for more context.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Condition">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>conditions represents the observations of a LXCMachine&rsquo;s current state.
Known condition types are Ready, InstanceProvisioned, Deleting, Paused.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="infrastructure.cluster.x-k8s.io/v1alpha2.SecretRef">SecretRef
</h3>
<p>
(<em>Appears on:</em>
<a href="#infrastructure.cluster.x-k8s.io/v1alpha2.LXCClusterSpec">LXCClusterSpec</a>)
</p>
<p>
<p>SecretRef is a reference to a secret in the cluster.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the secret to use. The secret must already exist in the same namespace as the parent object.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>.
</em></p>
