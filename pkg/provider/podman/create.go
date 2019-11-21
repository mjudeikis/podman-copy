package podman

import (
	"context"

	"github.com/virtual-kubelet/virtual-kubelet/log"
	v1 "k8s.io/api/core/v1"
)

// CreatePod accepts a Pod definition and stores it in memory.
func (p *PodmanV0Provider) CreatePod(ctx context.Context, pod *v1.Pod) error {
	log.G(ctx).Infof("receive CreatePod %q", pod.Name)
	err := p.c.Create(ctx, pod)
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
