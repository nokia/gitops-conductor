package controller

import (
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	CRDPlural   string = "customresourcedefinitions"
	CRDGroup    string = "apiextensions.k8s.io"
	CRDVersion  string = "v1beta1"
	FullCRDName string = CRDPlural + "." + CRDGroup
)

// Create a  Rest client with the new CRD Schema
var CrdSchemeGroupVersion = schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}

func AddToScheme(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(CrdSchemeGroupVersion,
		&extensionsobj.CustomResourceDefinition{},
	)
	meta_v1.AddToGroupVersion(scheme, CrdSchemeGroupVersion)
	return nil
}
