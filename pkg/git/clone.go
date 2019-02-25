package git

import (
	"fmt"
	"os"
	"strings"

	opsv1alpha1 "github.com/nokia/gitops-conductor/pkg/apis/ops/v1alpha1"
	gitc "gopkg.in/src-d/go-git.v4"
	"k8s.io/apimachinery/pkg/util/rand"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("git")

var baseDir = "/data/"

func SetupGit(spec *opsv1alpha1.GitOps) (string, error) {

	st := rand.String(10)
	cloneDir := ""
	if spec.Status.RootFolder != "" && strings.Contains(spec.Status.RootFolder, baseDir) {
		cloneDir = spec.Status.RootFolder
	} else {
		cloneDir = baseDir + st
	}
	err := os.Mkdir(cloneDir, 0755)
	if err != nil && !os.IsExist(err) {
		log.Error(err, "Failed to create directory")
		return "", err
	}

	url := ""

	if spec.Spec.User != "" {
		log.Info("Cloning with authentication")
		url = fmt.Sprintf("https://%s:%s@%s", spec.Spec.User, spec.Spec.Password, spec.Spec.Repository)
	} else {
		url = fmt.Sprintf("https://%s", spec.Spec.Repository)
	}
	cloneOpt := &gitc.CloneOptions{
		URL: url,
	}
	_, err = gitc.PlainClone(cloneDir, false, cloneOpt)
	if err != nil {
		log.Error(err, "Failed to clone repo:", "user", spec.Spec.User)
		return "", err
	}

	return cloneDir, nil
}
