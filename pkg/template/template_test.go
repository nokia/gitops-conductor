package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	opsv1alpha1 "github.com/nokia/gitops-conductor/pkg/apis/ops/v1alpha1"
	"github.com/nokia/gitops-conductor/tests/utils"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func TestTemplate(t *testing.T) {

	logf.SetLogger(&utils.Logg{})
	ops := utils.GetDefaultOps()
	ops.Status.RootFolder = "../../tests/yamls"
	ops.Spec.Templating = &opsv1alpha1.Templating{
		Enabled: true,
		Executor: &opsv1alpha1.Executor{
			Exec: "sed",
			Args: []string{
				`s/TEST/DONE/g`,
			},
		},
	}
	err := RunGoTemplate(ops)
	assert.Nil(t, err, "Templating failed")
}

func TestRunPreExec(t *testing.T) {

	ops := &opsv1alpha1.GitOps{}
	ops.Status.RootFolder = "../../tests/data/app"
	ops.Spec.Templating = &opsv1alpha1.Templating{
		Enabled: true,
		Executor: &opsv1alpha1.Executor{
			Exec: "sed",
			Args: []string{
				`s/TEST/DONE/g`,
			},
		},
	}

	err := RunPreExecutor(ops, "")
	assert.Nil(t, err)

}
