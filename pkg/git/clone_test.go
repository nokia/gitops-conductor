package git

import (
	"testing"

	"github.com/nokia/gitops-conductor/tests/utils"
	"github.com/stretchr/testify/assert"
	gitc "gopkg.in/src-d/go-git.v4"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func TestClonePullBranch(t *testing.T) {
	logf.SetLogger(&utils.Logg{})
	spec := utils.GetDefaultOps()
	spec.Spec.Branch = "master"
	baseDir = "/tmp/data"
	utils.CreateBaseDir(baseDir)
	rootDir, err := SetupGit(spec)
	spec.Status.RootFolder = rootDir
	assert.Nil(t, err, "Setup Git failed")
	assert.NotEqual(t, rootDir, "", "Setup git failed root dir not create")

	err = Pull(spec)
	assert.Equal(t, err, gitc.NoErrAlreadyUpToDate)

	err = Pull(spec)
	assert.Equal(t, err, gitc.NoErrAlreadyUpToDate, "Pull branch failed")
}
