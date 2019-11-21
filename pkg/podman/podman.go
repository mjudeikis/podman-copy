package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/varlink/go/varlink"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/virtual-kubelet/podman/pkg/converter"
	"github.com/virtual-kubelet/podman/pkg/iopodman"
	"github.com/virtual-kubelet/podman/pkg/util/errors"
)

var (
	// Provider configuration defaults.
	defaultSocket = "unix:/run/podman/io.podman"
	defaultSleep  = time.Millisecond * 100
)

// Config defines podman configurables
type Config struct {
	Socket *string
	Log    *zap.SugaredLogger
}

type conn struct {
	varlink.Connection
	sync.Mutex
}

type podman struct {
	c   *conn
	log *zap.SugaredLogger
}

// Podman is an simplified interface to interfact with
// podman varlink api
type Podman interface {
	// Methdods - locking the connection
	Create(ctx context.Context, pod *corev1.Pod) error
	Delete(ctx context.Context, pod *corev1.Pod) error
	GetByName(ctx context.Context, name string) (*corev1.Pod, error)
	List(ctx context.Context) (*corev1.PodList, error)
	// Methods - using above methods
	Update(ctx context.Context, pod *corev1.Pod) error
	CreateOrUpdate(ctx context.Context, pod *corev1.Pod) error
	Get(ctx context.Context, pod *corev1.Pod) (*corev1.Pod, error)
}

// New created new instance of podman interface
func New(ctx context.Context, c *Config) (Podman, error) {
	podman := podman{}
	cfg := getConfig(c)
	var err error
	vConn, err := varlink.NewConnection(ctx, *cfg.Socket)
	if err != nil {
		return nil, err
	}

	conn := conn{
		Connection: *vConn,
	}

	podman.c = &conn
	podman.log = cfg.Log

	return podman, nil
}

func getConfig(c *Config) *Config {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	log := logger.Sugar()

	if c != nil {
		if c.Socket == nil {
			c.Socket = &defaultSocket
		}
		if c.Log == nil {
			c.Log = log
		}
		return c
	}

	return &Config{
		Socket: &defaultSocket,
		Log:    log,
	}
}

// Create creates podman pod and containers within
func (p podman) Create(ctx context.Context, pod *corev1.Pod) error {
	if pod == nil {
		return fmt.Errorf("create pod can't be nil")
	}

	key := converter.BuildKey(pod)
	podmanPod, err := converter.GetPodmanPod(key, pod)
	if err != nil {
		p.log.Error("getPodmanPod failed", "err", err.Error())
		return err
	}
	p.c.Lock()
	podmanPodName, err := iopodman.CreatePod().Call(ctx, &p.c.Connection, *podmanPod)
	p.c.Unlock()
	if err != nil {
		p.log.Error("create pod failed", "err", err.Error())
		return errors.VKError(err)
	}
	p.log.Info("pod created ", "podName ", podmanPodName)

	// add containers in the pod
	for _, c := range pod.Spec.Containers {
		p.log.Info("create container ", "pod ", podmanPodName, " container ", c.Name)
		container := converter.KubeSpecToPodmanContainer(c, podmanPodName)
		p.c.Lock()
		_, err := iopodman.CreateContainer().Call(ctx, &p.c.Connection, container)
		p.c.Unlock()
		if err != nil {
			p.log.Error("error createContainer", "err", err.Error())
			return errors.VKError(err)
		}
	}

	// start pod
	p.c.Lock()
	_, err = iopodman.StartPod().Call(ctx, &p.c.Connection, podmanPodName)
	p.c.Unlock()
	if err != nil {
		p.log.Error("error startPod", "err", err.Error())
		return errors.VKError(err)
	}

	// check pod status
	retry := 1
	for retry < 5 {
		retry++
		p.c.Lock()
		podmanPodStatus, err := iopodman.InspectPod().Call(ctx, &p.c.Connection, podmanPodName)
		p.c.Unlock()
		if err != nil {
			p.log.Error("error GetPod.InspectPod ", "err ", err.Error())
			return errors.VKError(err)
		}

		var status PodmanPod
		err = json.Unmarshal([]byte(podmanPodStatus), &status)
		if err != nil {
			return errors.VKError(err)
		}
		healty := true
		for _, c := range status.Containers {
			if c.State != "running" {
				healty = false
			}
		}
		if healty {
			continue
		}
	}

	return nil
}

func (p podman) CreateOrUpdate(ctx context.Context, pod *corev1.Pod) error {
	if pod == nil {
		return fmt.Errorf("create pod can't be nil")
	}

	// for logging only
	key := converter.BuildKey(pod)

	pp, err := p.Get(ctx, pod)
	if err != nil {
		if _, ok := err.(*iopodman.PodNotFound); ok {
			p.log.Debugf("pod not found, creating", " pod ", key)
			return p.Create(ctx, pod)
		}
		if pp != nil && err == nil {
			p.log.Debugf("pod exist, update", " pod ", key)
			return p.Update(ctx, pod)
		}
	}

	return nil
}

func (p podman) Delete(ctx context.Context, pod *corev1.Pod) error {
	if pod == nil {
		p.log.Error("pod can't be nil")
		return fmt.Errorf("pod can't be nil")
	}

	key := converter.BuildKey(pod)
	p.c.Lock()
	_, err := iopodman.RemovePod().Call(ctx, &p.c.Connection, key, true)
	p.c.Unlock()
	if err != nil {
		p.log.Error("error while deleting pod", " pod ", key, " err ", err.Error())
		return errors.VKError(err)
	}

	return nil
}

func (p podman) Update(ctx context.Context, pod *corev1.Pod) error {
	err := p.Delete(ctx, pod)
	if err != nil {
		return errors.VKError(err)
	}
	return p.Create(ctx, pod)
}

func (p podman) Get(ctx context.Context, input *corev1.Pod) (pod *v1.Pod, err error) {
	key := converter.BuildKey(input)
	return p.GetByName(ctx, key)
}

func (p podman) GetByName(ctx context.Context, name string) (pod *v1.Pod, err error) {
	p.c.Lock()
	_, err = iopodman.GetPod().Call(ctx, &p.c.Connection, name)
	p.c.Unlock()
	if err != nil {
		return nil, errors.VKError(err)
	}

	p.c.Lock()
	ppod, err := iopodman.InspectPod().Call(ctx, &p.c.Connection, name)
	p.c.Unlock()
	if err != nil {
		return nil, err
	}

	if len(ppod) > 0 {
		kpod, err := converter.GetKubePod(ppod)
		if err != nil {
			return nil, errors.VKError(err)
		}
		return kpod, nil
	}
	return nil, errors.VKError(err)

}

func (p podman) List(ctx context.Context) (podList *corev1.PodList, err error) {
	p.c.Lock()
	ppods, err := iopodman.ListPods().Call(ctx, &p.c.Connection)
	p.c.Unlock()
	if err != nil {
		return nil, errors.VKError(err)
	}

	kpodsList := &corev1.PodList{}
	for _, podData := range ppods {
		kpod, err := p.GetByName(ctx, podData.Name)
		if err != nil {
			return nil, errors.VKError(err)
		}
		kpodsList.Items = append(kpodsList.Items, *kpod)
	}

	return kpodsList, nil
}
