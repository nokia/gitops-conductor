package template

import (
	"io/ioutil"

	opsv1alpha1 "github.com/nokia/gitops-conductor/pkg/apis/ops/v1alpha1"
	"gopkg.in/yaml.v2"
)

func IsBlacklisted(spec *opsv1alpha1.GitOps) bool {
	if spec.Spec.Templating != nil {
		if spec.Spec.Templating.Source != nil {
			if spec.Spec.Templating.Source.BlackListFile != "" {
				sourceFile := spec.Status.RootFolder + "/" + spec.Spec.RootFolder + "/" + spec.Spec.Templating.SourceFolder + "/" + spec.Spec.Templating.Source.BlackListFile
				data, err := ioutil.ReadFile(sourceFile)
				if err != nil {
					log.Error(err, "Failed to read Blacklist file")
					return false
				}

				d := opsv1alpha1.BlacklistContent{}
				yaml.Unmarshal(data, &d)

				dataMap, err := getTemplatingSource(spec)
				if err != nil {
					log.Error(err, "Failed to get template source")
					return false
				}
				if val, ok := dataMap[d.Identifier]; ok {
					for _, v := range d.Values {
						if v == val {
							return true
						}
					}
				}
			}
		}
	}

	return false
}
