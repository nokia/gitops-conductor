package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	opsv1alpha1 "github.com/nokia/gitops-conductor/pkg/apis/ops/v1alpha1"
	"github.com/nokia/gitops-conductor/pkg/template"
	gitc "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

func Pull(spec *opsv1alpha1.GitOps) error {

	r, err := gitc.PlainOpen(spec.Status.RootFolder)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	pullOpt := &gitc.PullOptions{RemoteName: "origin", Force: true}
	err = w.Pull(pullOpt)
	if err != nil && err != gitc.NoErrAlreadyUpToDate {
		log.Error(err, "Pull failed")
		return err
	}

	return err
}

func CheckoutBranch(spec *opsv1alpha1.GitOps) error {
	branch := spec.Spec.Branch
	if branch == "" {
		branch = "master"
	}
	r, err := gitc.PlainOpen(spec.Status.RootFolder)
	if err != nil {
		return err
	}
	err = r.Fetch(&gitc.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*"},
		Force:    true,
	})
	if err != nil && err != gitc.NoErrAlreadyUpToDate {
		log.Error(err, "Failed to fetch refs", "Branch", branch)
		return err
	} else if err != nil && err == gitc.NoErrAlreadyUpToDate {
		checkBranch(r, branch)
		return err
	}
	return checkBranch(r, branch)
}

func checkBranch(r *gitc.Repository, branch string) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	head, err := r.Head()
	if head.Name() == plumbing.ReferenceName(branch) {
		//Already on correct branch
		return gitc.NoErrAlreadyUpToDate
	} else {
		err = w.Checkout(&gitc.CheckoutOptions{
			Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
			Create: false,
			Force:  true,
		})
		if err != nil {
			log.Error(err, "Failed to checkout branch", "Branch", branch)
			return err
		}
	}
	//Reset to HEAD
	err = w.Reset(&gitc.ResetOptions{Mode: gitc.HardReset})
	if err != nil {
		log.Error(err, "Failed reset", "Branch", branch)
		return err
	}
	return nil
}

//PullTemplates fetches the yaml templates from specified directory. Runs pre-templating if specified by the CRD
func PullTemplates(spec *opsv1alpha1.GitOps, dir string, scheme *runtime.Scheme) ([]runtime.Object, error, int) {

	objs := []runtime.Object{}

	tmplRoot := dir + spec.Status.RootFolder + "/" + spec.Spec.RootFolder

	if spec.Spec.Templating != nil {

		err := template.RunPreExecutor(spec, "")
		if err != nil {
			log.Error(err, "Failed to run preTemplating")
			return []runtime.Object{}, err, 0
		}
		if template.IsBlacklisted(spec) {
			return objs, nil, 0
		}
		err = template.RunGoTemplate(spec)
		if err != nil {
			log.Error(err, "Failed to run templating ")
			return []runtime.Object{}, err, 0
		}
		tmplRoot = dir + spec.Status.RootFolder + "/_output"
	}

	objs, inv := parseFolder(tmplRoot, scheme)

	return objs, nil, inv
}

func parseFile(tmplRoot string, file os.FileInfo, scheme *runtime.Scheme) ([]runtime.Object, int) {
	decoder := serializer.NewCodecFactory(scheme)
	decode := decoder.UniversalDeserializer()
	objects := []runtime.Object{}
	invalid := 0
	if strings.Contains(file.Name(), ".yaml") {

		data, err := ioutil.ReadFile(tmplRoot + "/" + file.Name())
		if err != nil {
			log.Error(err, "Failed to read file")
			return objects, 1
		}
		individual := strings.Split(string(data), "\n---")
		for _, o := range individual {
			obj, _, err := decode.Decode([]byte(o), nil, nil)
			if err != nil {
				log.Error(err, "Failed to parse file", "FileName", file.Name())
				invalid += 1
			}
			m, err := meta.Accessor(obj)
			if err != nil {
				continue
			}
			if m.GetNamespace() == "" {
				m.SetNamespace("default")
			}
			log.Info("Obj", "Obj", obj)
			objects = append(objects, obj)
		}
	}
	return objects, invalid
}

func parseFolder(tmplRoot string, scheme *runtime.Scheme) ([]runtime.Object, int) {
	objs := []runtime.Object{}
	files, err := ioutil.ReadDir(tmplRoot)
	if err != nil {
		log.Error(err, "Failed to read dir")
	}
	invalid := 0
	for _, file := range files {
		var newObjs []runtime.Object
		var inv int
		if file.IsDir() {
			newObjs, inv = parseFolder(tmplRoot+"/"+file.Name(), scheme)
		} else {
			newObjs, inv = parseFile(tmplRoot, file, scheme)
		}
		for _, o := range newObjs {
			objs = append(objs, o)
		}
		invalid += inv
	}
	return objs, invalid
}
