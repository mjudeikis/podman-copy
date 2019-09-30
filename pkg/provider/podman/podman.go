package podman

import (
	"time"

	"github.com/varlink/go/varlink"
	v1 "k8s.io/api/core/v1"

	"github.com/virtual-kubelet/podman/pkg/store"
)

const (
	// Provider configuration defaults.
	defaultCPUCapacity       = "5"
	defaultMemoryCapacity    = "2Gi"
	defaultPodCapacity       = "10"
	defaultSocket            = "unix:/run/podman/io.podman"
	defaultDaemonSetDisabled = "true"
	defaultStorageDir        = "/usr/local/vk-pods"
)

// PodmanV0Provider implements the virtual-kubelet provider interface and stores pods in memory.
type PodmanV0Provider struct { //nolint:golint
	nodeName           string
	operatingSystem    string
	pods               map[string]*v1.Pod
	config             PodmanConfig
	startTime          time.Time
	notifier           func(*v1.Pod)
	internalIP         string
	daemonEndpointPort int32
	connection         varlink.Connection

	// we need local storage until we can query pods by default
	// TODO: Store k8s pod spec into podman metadata so we can remove this
	store store.Store
}

// PodmanProvider is like PodmanV0Provider, but implements the PodNotifier interface
type PodmanProvider struct { //nolint:golint
	*PodmanV0Provider
}

// PodmanConfig contains a podman virtual-kubelet's configurable parameters.
type PodmanConfig struct { //nolint:golint
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
	Pods   string `json:"pods,omitempty"`

	Socket string `json:"socket,omitempty"`

	DaemonSetDisabled string `json:"daemonSetDisabled,omitempty"`
	StorageLocation   string `json:"storageLocation,omitempty"`
}

// NewPodmanProviderPodmanConfig creates a new PodmanV0Provider. podman legacy provider does not implement the new asynchronous podnotifier interface
func NewPodmanV0ProviderPodmanConfig(config PodmanConfig, nodeName, operatingSystem string) (*PodmanV0Provider, error) {
	c, err := varlink.NewConnection(config.Socket)
	if err != nil {
		return nil, err
	}

	storage := store.New(config.StorageLocation)

	provider := PodmanV0Provider{
		nodeName:        nodeName,
		operatingSystem: operatingSystem,
		pods:            make(map[string]*v1.Pod),
		config:          config,
		startTime:       time.Now(),
		connection:      *c,
		store:           storage,
		// By default notifier is set to a function which is a no-op. In the event we've implemented the PodNotifier interface,
		// it will be set, and then we'll call a real underlying implementation.
		// This makes it easier in the sense we don't need to wrap each method.
		notifier: func(*v1.Pod) {},
	}

	return &provider, nil
}

// NewPodmanV0Provider creates a new PodmanV0Provider
func NewPodmanV0Provider(providerConfig, nodeName, operatingSystem string) (*PodmanV0Provider, error) {
	config, err := loadConfig(providerConfig, nodeName)
	if err != nil {
		return nil, err
	}

	return NewPodmanV0ProviderPodmanConfig(config, nodeName, operatingSystem)
}

// NewPodmanProviderPodmanConfig creates a new PodmanProvider with the given config
func NewPodmanProviderPodmanConfig(config PodmanConfig, nodeName, operatingSystem string) (*PodmanProvider, error) {
	p, err := NewPodmanV0ProviderPodmanConfig(config, nodeName, operatingSystem)

	return &PodmanProvider{PodmanV0Provider: p}, err
}

// NewPodmanProvider creates a new PodmanProvider, which implements the PodNotifier interface
func NewPodmanProvider(providerConfig, nodeName, operatingSystem string) (*PodmanProvider, error) {
	config, err := loadConfig(providerConfig, nodeName)
	if err != nil {
		return nil, err
	}

	return NewPodmanProviderPodmanConfig(config, nodeName, operatingSystem)
}
