package podman

import "time"

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
