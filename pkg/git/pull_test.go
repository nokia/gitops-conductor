package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nokia/gitops-conductor/tests/utils"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func TestPullYaml(t *testing.T) {

	logf.SetLogger(&utils.Logg{})
	ops := utils.GetDefaultOps()
	ops.Status.RootFolder = "../../tests/yamls"
	baseDir = "/tmp"

	c, _ := config.GetConfig()
	mgr, err := manager.New(c, manager.Options{})
	assert.Nil(t, err, "Failed to get manager")
	obj, err := PullTemplates(ops, "", mgr.GetScheme())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(obj))
	t.Logf("Test: %v ", obj)
}
