package crd

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("crd")

type CrdConfig struct {
	Crds []struct {
		Group   string `yaml:"group"`
		Version string `yaml:"version"`
		Kind    string `yaml:"kind"`
	} `yaml:"crds"`
}

func addKnownTypes(scheme *runtime.Scheme, gvk schema.GroupVersionKind) error {
	log.Info("Registering CRD", "Kind", gvk.Kind, "Group", gvk.Group)
	typ := &unstructured.Unstructured{}
	typ.SetGroupVersionKind(gvk)
	scheme.AddKnownTypeWithName(gvk,
		typ,
	)
	metav1.AddToGroupVersion(scheme, gvk.GroupVersion())
	return nil
}

func AddKnowCrds(scheme *runtime.Scheme, file string) error {
	log.Info("Creating Additional resources to scheme")
	c, err := readConfig(file)
	if err != nil {
		return err
	}
	for _, v := range c.Crds {
		gvk := schema.GroupVersionKind{Group: v.Group, Version: v.Version, Kind: v.Kind}
		addKnownTypes(scheme, gvk)
	}
	return nil
}

func readConfig(file string) (*CrdConfig, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error(err, "Failed to register new types")
		return nil, err
	}
	c := &CrdConfig{}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		log.Error(err, "Failed to unmarshal yaml")
		return nil, err
	}
	return c, nil
}
