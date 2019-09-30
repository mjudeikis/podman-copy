package podman

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/virtual-kubelet/virtual-kubelet/log"
	"github.com/virtual-kubelet/virtual-kubelet/trace"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/virtual-kubelet/podman/pkg/iopodman"
)

// CreatePod accepts a Pod definition and stores it in memory.
func (p *PodmanV0Provider) CreatePod(ctx context.Context, pod *v1.Pod) error {
	ctx, span := trace.StartSpan(ctx, "CreatePod")
	defer span.End()

	// Add the pod's coordinates to the current span.
	ctx = addAttributes(ctx, span, namespaceKey, pod.Namespace, nameKey, pod.Name)

	log.G(ctx).Infof("receive CreatePod %q", pod.Name)

	key, err := buildKey(pod)
	if err != nil {
		return err
	}

	// check if DS
	ds, _ := strconv.ParseBool(p.config.DaemonSetDisabled)
	if ds {
		for _, o := range pod.GetObjectMeta().GetOwnerReferences() {
			if strings.EqualFold(o.Kind, "DaemonSet") {
				return p.noOpPod(pod)

			}
		}
	}

	podmanPod := iopodman.PodCreate{
		Name:   key,
		Labels: pod.Labels,
	}

	result, err := iopodman.CreatePod().Call(&p.connection, podmanPod)
	if err != nil {
		return err
	}
	log.G(ctx).Infof("store pod def to %s", result)
	return p.store.Put(result, pod)
}

func (p *PodmanV0Provider) noOpPod(pod *v1.Pod) error {

	key, err := buildKey(pod)
	if err != nil {
		return err
	}

	now := metav1.NewTime(time.Now())
	pod.Status = v1.PodStatus{
		Phase:     v1.PodRunning,
		HostIP:    "0.0.0.0",
		PodIP:     "0.0.0.0",
		StartTime: &now,
		Conditions: []v1.PodCondition{
			{
				Type:    v1.PodInitialized,
				Status:  v1.ConditionTrue,
				Message: "DaemonSet scheduling disabled. NoOp pod.",
			},
		},
	}

	for _, container := range pod.Spec.Containers {
		pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, v1.ContainerStatus{
			Name:         container.Name,
			Image:        container.Image,
			Ready:        true,
			RestartCount: 0,
			State: v1.ContainerState{
				Running: &v1.ContainerStateRunning{
					StartedAt: now,
				},
			},
		})
	}

	p.pods[key] = pod
	p.notifier(pod)
	return nil
}
