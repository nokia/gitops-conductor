package template

import (
	"io/ioutil"
	"path"

	"text/template"

	"github.com/Masterminds/sprig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func templateFuncs(baseDir string) template.FuncMap {
	m := sprig.TxtFuncMap()

	m["insertFile"] = func(file string) (string, error) {
		data, err := ioutil.ReadFile(path.Join(baseDir, file))
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	m["k8sNodes"] = func() (int, error) {
		config, err := rest.InClusterConfig()
		if err != nil {
			return 0, err
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return 0, err
		}
		nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return 0, err
		}
		return len(nodes.Items), nil
	}
	return m

}
