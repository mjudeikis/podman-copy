package podman

import (
	"context"

	"github.com/virtual-kubelet/virtual-kubelet/log"
	"github.com/virtual-kubelet/virtual-kubelet/trace"
	v1 "k8s.io/api/core/v1"

	"github.com/virtual-kubelet/podman/pkg/iopodman"
)

// DeletePod deletes the specified pod out of memory.
func (p *PodmanV0Provider) DeletePod(ctx context.Context, pod *v1.Pod) (err error) {
	ctx, span := trace.StartSpan(ctx, "DeletePod")
	defer span.End()

	// Add the pod's coordinates to the current span.
	ctx = addAttributes(ctx, span, namespaceKey, pod.Namespace, nameKey, pod.Name)

	log.G(ctx).Infof("receive DeletePod %q", pod.Name)

	key, err := buildKey(pod)
	if err != nil {
		return err
	}

	// TODO: Check if exist
	result, err := iopodman.StopPod().Call(&p.connection, key, 0)
	if err != nil {
		return err
	}
	log.G(ctx).Info(result)

	p.notifier(pod)

	return nil
}
