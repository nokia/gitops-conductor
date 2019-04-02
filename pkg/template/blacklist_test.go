package template

import (
	"testing"

	opsv1alpha1 "github.com/nokia/gitops-conductor/pkg/apis/ops/v1alpha1"
	"github.com/nokia/gitops-conductor/tests/utils"
	"github.com/stretchr/testify/assert"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func TestBlacklist(t *testing.T) {
	logf.SetLogger(&utils.Logg{})
	ops := utils.GetDefaultOps()
	ops.Status.RootFolder = "../../tests/yamls"
	ops.Spec.Templating = &opsv1alpha1.Templating{
		Enabled:      true,
		SourceFolder: "source",
		Source: &opsv1alpha1.TemplateDataSource{
			TemplateDataFile: []string{"template.yaml"},
			BlackListFile:    "blacklist.yaml",
		},
		Executor: &opsv1alpha1.Executor{
			Exec: "sed",
			Args: []string{
				`s/TEST/DONE/g`,
			},
		},
	}
	isBlack := IsBlacklisted(ops)
	assert.True(t, isBlack)

}
