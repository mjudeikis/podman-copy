package converter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/virtual-kubelet/podman/pkg/iopodman"
)

func BuildKeyFromNames(namespace string, name string) (string, error) {
	return fmt.Sprintf("%s-%s", namespace, name), nil
}

// BuildKey is a helper for building the "key" for the providers pod store.
func BuildKey(pod *v1.Pod) string {
	if pod.Namespace == "" {
		return pod.Name
	}
	return fmt.Sprintf("%s-%s", pod.Namespace, pod.Name)
}

func SplitPodName(key string) (namespace, name string) {
	keys := strings.Split(key, "-")
	return keys[0], keys[1]
}

func KubeSpecToPodmanContainer(container v1.Container, podName string) iopodman.Create {
	// TODO: Extend this to match most of the fields
	var args []string
	args = append(args, container.Image)
	args = append(args, container.Command...)
	args = append(args, container.Args...)
	containerName := fmt.Sprintf("%s-%s", podName, container.Name)
	return iopodman.Create{
		Args:    args,
		Command: &container.Command,
		Name:    &containerName,
		Pod:     &podName,
	}
}

// GetPodmanPod return podman pod with pod metadata in the label
func GetPodmanPod(key string, p *v1.Pod) (*iopodman.PodCreate, error) {
	// preserve original pod spec into lables
	pod := p.DeepCopy()
	data, err := yaml.Marshal(pod)
	if err != nil {
		return nil, err
	}
	podSpecBase := base64.StdEncoding.EncodeToString(data)
	if pod.Labels == nil {
		pod.Labels = make(map[string]string, 1)
	}
	pod.Labels["pod"] = podSpecBase

	podmanPod := iopodman.PodCreate{
		Name:   key,
		Labels: pod.Labels,
	}

	return &podmanPod, nil
}

type PodmanPod struct {
	Config struct {
		ID           string            `json:"id"`
		Name         string            `json:"name"`
		Labels       map[string]string `json:"labels"`
		CgroupParent string            `json:"cgroupParent"`
		SharesCgroup bool              `json:"sharesCgroup"`
		InfraConfig  struct {
			MakeInfraContainer bool        `json:"makeInfraContainer"`
			InfraPortBindings  interface{} `json:"infraPortBindings"`
		} `json:"infraConfig"`
		Created time.Time `json:"created"`
		LockID  int       `json:"lockID"`
	} `json:"Config"`
	State struct {
		CgroupPath       string `json:"cgroupPath"`
		InfraContainerID string `json:"infraContainerID"`
	} `json:"State"`
	Containers []struct {
		ID    string `json:"id"`
		State string `json:"state"`
	} `json:"Containers"`
}

func GetKubePod(ppodJson string) (*v1.Pod, error) {
	var ppod PodmanPod
	err := json.Unmarshal([]byte(ppodJson), &ppod)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(ppod.Config.Labels["pod"])
	if err != nil {
		return nil, err
	}
	var kpod v1.Pod
	err = yaml.Unmarshal(data, &kpod)
	if err != nil {
		return nil, err
	}

	// configure status for the kubePod
	kpod.Status, err = GetPodStatus(ppod)
	if err != nil {
		return nil, err
	}

	return &kpod, nil
}

func GetPodStatus(ppod PodmanPod) (v1.PodStatus, error) {
	now := metav1.NewTime(time.Now())
	status := v1.PodStatus{}
	status.StartTime = &now
	status.HostIP = "1.2.3.4"
	status.PodIP = "5.6.7.8"
	status.Conditions = []v1.PodCondition{
		{
			Type:   v1.PodInitialized,
			Status: v1.ConditionTrue,
		},
		{
			Type:   v1.PodReady,
			Status: v1.ConditionTrue,
		},
		{
			Type:   v1.PodScheduled,
			Status: v1.ConditionTrue,
		},
	}

	for _, c := range ppod.Containers {
		containerStatus := v1.ContainerStatus{}
		containerStatus.Name = c.ID
		containerStatus.Image = c.ID
		var state v1.ContainerState
		switch c.State {
		case "running":
			state = v1.ContainerState{
				Running: &v1.ContainerStateRunning{
					StartedAt: metav1.Time{
						Time: ppod.Config.Created,
					},
				},
			}
			containerStatus.Ready = true
			status.Phase = v1.PodRunning
		case "exited":
			state = v1.ContainerState{
				Terminated: &v1.ContainerStateTerminated{
					StartedAt: metav1.Time{
						Time: time.Now(),
					},
				},
			}
			status.Phase = v1.PodFailed
			containerStatus.Ready = false
		default:
			state = v1.ContainerState{
				Terminated: &v1.ContainerStateTerminated{
					StartedAt: metav1.Time{
						Time: time.Now(),
					},
				},
			}
			status.Phase = v1.PodUnknown
			containerStatus.Ready = false
		}
		containerStatus.State = state
		status.ContainerStatuses = append(status.ContainerStatuses, containerStatus)
	}

	return status, nil
}
