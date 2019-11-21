package podman

import (
	"context"

	"github.com/virtual-kubelet/virtual-kubelet/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	stats "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
)

const (
	// Values used in tracing as attribute keys.
	namespaceKey     = "namespace"
	nameKey          = "name"
	containerNameKey = "containerName"
)

// UpdatePod accepts a Pod definition and updates its reference.
func (p *PodmanV0Provider) UpdatePod(ctx context.Context, pod *v1.Pod) error {
	log.G(ctx).Infof("receive UpdatePod %q", pod.Name)
	err := p.c.Update(ctx, pod)
	if err != nil {
		return err
	}

	pod, err = p.c.Get(ctx, pod)
	if err != nil {
		return err
	}

	p.notifier(pod)
	return nil
}

func (p *PodmanV0Provider) ConfigureNode(ctx context.Context, n *v1.Node) {
	n.Status.Capacity = p.capacity()
	n.Status.Allocatable = p.capacity()
	n.Status.Conditions = p.nodeConditions()
	n.Status.Addresses = p.nodeAddresses()
	n.Status.DaemonEndpoints = p.nodeDaemonEndpoints()
	os := p.operatingSystem
	if os == "" {
		os = "Linux"
	}
	n.Status.NodeInfo.OperatingSystem = os
	n.Status.NodeInfo.Architecture = "amd64"
	n.ObjectMeta.Labels["alpha.service-controller.kubernetes.io/exclude-balancer"] = "true"
}

// Capacity returns a resource list containing the capacity limits.
func (p *PodmanV0Provider) capacity() v1.ResourceList {
	return v1.ResourceList{
		"cpu":    resource.MustParse(p.config.CPU),
		"memory": resource.MustParse(p.config.Memory),
		"pods":   resource.MustParse(p.config.Pods),
	}
}

// NodeConditions returns a list of conditions (Ready, OutOfDisk, etc), for updates to the node status
// within Kubernetes.
func (p *PodmanV0Provider) nodeConditions() []v1.NodeCondition {
	// TODO: Make this configurable
	return []v1.NodeCondition{
		{
			Type:               "Ready",
			Status:             v1.ConditionTrue,
			LastHeartbeatTime:  metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "KubeletReady",
			Message:            "kubelet is ready.",
		},
		{
			Type:               "OutOfDisk",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "KubeletHasSufficientDisk",
			Message:            "kubelet has sufficient disk space available",
		},
		{
			Type:               "MemoryPressure",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "KubeletHasSufficientMemory",
			Message:            "kubelet has sufficient memory available",
		},
		{
			Type:               "DiskPressure",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "KubeletHasNoDiskPressure",
			Message:            "kubelet has no disk pressure",
		},
		{
			Type:               "NetworkUnavailable",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "RouteCreated",
			Message:            "RouteController created a route",
		},
	}

}

// NodeAddresses returns a list of addresses for the node status
// within Kubernetes.
func (p *PodmanV0Provider) nodeAddresses() []v1.NodeAddress {
	return []v1.NodeAddress{
		{
			Type:    "InternalIP",
			Address: p.internalIP,
		},
	}
}

// NodeDaemonEndpoints returns NodeDaemonEndpoints for the node status
// within Kubernetes.
func (p *PodmanV0Provider) nodeDaemonEndpoints() v1.NodeDaemonEndpoints {
	return v1.NodeDaemonEndpoints{
		KubeletEndpoint: v1.DaemonEndpoint{
			Port: p.daemonEndpointPort,
		},
	}
}

// GetStatsSummary returns dummy stats for all pods known by this provider.
func (p *PodmanV0Provider) GetStatsSummary(ctx context.Context) (*stats.Summary, error) {
	// Create the Summary object that will later be populated with node and pod stats.
	res := &stats.Summary{}

	// Populate the Summary object with basic node stats.
	res.Node = stats.NodeStats{
		NodeName:  p.nodeName,
		StartTime: metav1.NewTime(p.startTime),
	}

	// Populate the Summary object with dummy stats for each pod known by this provider.
	//for _, pod := range p.pods {
	//	var (
	//		// totalUsageNanoCores will be populated with the sum of the values of UsageNanoCores computes across all containers in the pod.
	//		totalUsageNanoCores uint64
	//		// totalUsageBytes will be populated with the sum of the values of UsageBytes computed across all containers in the pod.
	//		totalUsageBytes uint64
	//	)
	//
	//	// Create a PodStats object to populate with pod stats.
	//	pss := stats.PodStats{
	//		PodRef: stats.PodReference{
	//			Name:      pod.Name,
	//			Namespace: pod.Namespace,
	//			UID:       string(pod.UID),
	//		},
	//		StartTime: pod.CreationTimestamp,
	//	}
	//
	//	// Iterate over all containers in the current pod to compute dummy stats.
	//	for _, container := range pod.Spec.Containers {
	//		// Grab a dummy value to be used as the total CPU usage.
	//		// The value should fit a uint32 in order to avoid overflows later on when computing pod stats.
	//		dummyUsageNanoCores := uint64(rand.Uint32())
	//		totalUsageNanoCores += dummyUsageNanoCores
	//		// Create a dummy value to be used as the total RAM usage.
	//		// The value should fit a uint32 in order to avoid overflows later on when computing pod stats.
	//		dummyUsageBytes := uint64(rand.Uint32())
	//		totalUsageBytes += dummyUsageBytes
	//		// Append a ContainerStats object containing the dummy stats to the PodStats object.
	//		pss.Containers = append(pss.Containers, stats.ContainerStats{
	//			Name:      container.Name,
	//			StartTime: pod.CreationTimestamp,
	//			CPU: &stats.CPUStats{
	//				Time:           time,
	//				UsageNanoCores: &dummyUsageNanoCores,
	//			},
	//			Memory: &stats.MemoryStats{
	//				Time:       time,
	//				UsageBytes: &dummyUsageBytes,
	//			},
	//		})
	//	}
	//
	//	// Populate the CPU and RAM stats for the pod and append the PodsStats object to the Summary object to be returned.
	//	pss.CPU = &stats.CPUStats{
	//		Time:           time,
	//		UsageNanoCores: &totalUsageNanoCores,
	//	}
	//	pss.Memory = &stats.MemoryStats{
	//		Time:       time,
	//		UsageBytes: &totalUsageBytes,
	//	}
	//	res.Pods = append(res.Pods, pss)
	//}
	//
	//// Return the dummy stats.
	return res, nil
}

// NotifyPods is called to set a pod notifier callback function. This should be called before any operations are done
// within the provider.
func (p *PodmanProvider) NotifyPods(ctx context.Context, notifier func(*v1.Pod)) {
	p.notifier = notifier
}
