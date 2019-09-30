package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"

	"github.com/virtual-kubelet/virtual-kubelet/log"
)

type Store interface {
	Put(key string, pod *v1.Pod) error
	Get(key string) (*v1.Pod, error)
	List() ([]*v1.Pod, error)
	Delete(key string) error
}

type storage struct {
	dir string
}

var _ Store = &storage{}

func New(dir string) Store {
	return &storage{
		dir: dir,
	}
}

func (s *storage) Put(key string, pod *v1.Pod) error {
	b, err := yaml.Marshal(pod)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("%s/%s", s.dir, key)
	log.L.Debugf("write %s", path)
	err = os.MkdirAll(path, 0750)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(path, "pod.yaml"), b, 0666)
}

func (s *storage) Get(key string) (pod *v1.Pod, err error) {
	path := fmt.Sprintf("%s/%s", s.dir, key)
	b, err := ioutil.ReadFile(filepath.Join(path, "pod.yaml"))
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, &pod)
	return
}

func (s *storage) List() (pods []*v1.Pod, err error) {
	podsFiles, err := ioutil.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	for _, f := range podsFiles {
		b, err := ioutil.ReadFile(f.Name())
		if err != nil {
			return nil, err
		}
		var pod v1.Pod
		err = yaml.Unmarshal(b, &pod)
		pods = append(pods, &pod)
	}

	return
}

func (s *storage) Delete(key string) error {
	path := fmt.Sprintf("%s/%s", s.dir, key)
	return os.RemoveAll(path)
}
