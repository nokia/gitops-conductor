package template

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	templ "text/template"

	opsv1alpha1 "github.com/nokia/gitops-conductor/pkg/apis/ops/v1alpha1"
	yaml "gopkg.in/yaml.v2"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type templater struct {
	baseDir string
}

var log = logf.Log.WithName("controller_template")

//getTemplatingSource fetches template source files and adds them to a data set to be used when running
//the template function
func getTemplatingSource(spec *opsv1alpha1.GitOps) (map[string]interface{}, error) {

	if spec.Spec.Templating != nil {
		sourceDir := spec.Status.RootFolder + "/" + spec.Spec.RootFolder + "/" + spec.Spec.Templating.SourceFolder
		if spec.Spec.Templating.SourceFolder != "" {
			sourceDir += "/"
		}
		t := spec.Spec.Templating
		if t.Enabled {
			if t.Source != nil {
				data := make(map[string]interface{})
				for _, f := range t.Source.TemplateDataFile {
					temp := make(map[string]interface{})
					fData, err := ioutil.ReadFile(sourceDir + f)
					if err != nil {
						log.Error(err, "Failed to read file")
						return nil, err
					}
					yaml.Unmarshal(fData, temp)
					for k, v := range temp {
						data[k] = v
					}
				}
				return data, nil
			}
		}
		return nil, nil
	}
	return nil, nil
}

//RunGoTemplate executes template
func RunGoTemplate(spec *opsv1alpha1.GitOps) error {

	log.Info("Running go templates")
	outDir := spec.Status.RootFolder + "/_output"
	workDir := GetGitRootDir(spec)
	t := templater{baseDir: workDir}
	d, err := getTemplatingSource(spec)
	if err != nil {
		return err
	}
	err = os.Mkdir(outDir, 0755)
	if err != nil && !os.IsExist(err) {
		log.Error(err, "Failed to create output dir")
		return err
	}
	files, err := ioutil.ReadDir(workDir)
	if err != nil {
		log.Error(err, "Failed to read dir")
	}

	for _, file := range files {
		if file.IsDir() {
			if file.Name() == spec.Spec.Templating.SourceFolder {
				continue
			}
			t.templateDir(workDir+"/"+file.Name(), outDir+"/"+file.Name(), d)
		} else {
			t.templateFile(workDir, outDir, file, d)
		}
	}
	return nil
}

//templateFile takes one yaml file as input and runs Go template engine on the file. Output is written to output directory
func (t *templater) templateFile(workDir string, outDir string, file os.FileInfo, d map[string]interface{}) {
	if strings.Contains(file.Name(), "yaml") {

		filePath := workDir + "/" + file.Name()
		tEx := templ.New(file.Name())
		tEx.Funcs(templateFuncs(workDir))
		tEx.ParseFiles(filePath)
		b := bytes.NewBuffer([]byte{})
		err := tEx.Execute(b, d)
		if err != nil {
			log.Error(err, "Failed to execute template")
		}
		newF, err := os.Create(outDir + "/" + file.Name())
		if err != nil {
			log.Error(err, "Failed to create file", "file", file.Name())
			return
		}
		newF.Write(b.Bytes())
		newF.Close()
	}
}

//templateDir runs through a directory recursively templating every file on the way down in the tree
func (t *templater) templateDir(workDir string, outDir string, d map[string]interface{}) {
	log.Info("templating", "dir", workDir)
	err := os.Mkdir(outDir, 0755)
	if err != nil && !os.IsExist(err) {
		log.Error(err, "Failed to create output dir")
		return
	}
	files, err := ioutil.ReadDir(workDir)
	if err != nil {
		log.Error(err, "Failed to read dir")
	}

	for _, file := range files {
		if file.IsDir() {
			t.templateDir(workDir+"/"+file.Name(), outDir+"/"+file.Name(), d)
		} else {
			t.templateFile(workDir, outDir, file, d)
		}
	}

}

//RunPreExecutor runs an arbitary pre-executor to template the yaml files
func RunPreExecutor(spec *opsv1alpha1.GitOps, dir string) error {
	if spec.Spec.Templating != nil {
		t := spec.Spec.Templating
		if t.Executor != nil {

			// Create a new context and add a timeout to it
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel() // The cancel should be deferred so resources are cleaned up
			cmd := exec.CommandContext(ctx, t.Executor.Exec)
			cmd.Dir = GetGitRootDir(spec)
			if spec.Spec.Templating.SourceFolder != "" {
				cmd.Dir += "/" + spec.Spec.Templating.SourceFolder
			}

			if len(t.Executor.Args) >= 1 {
				a := []string{t.Executor.Exec}
				for _, add := range t.Executor.Args {
					a = append(a, add)
				}
				cmd.Args = a
			}

			out, err := cmd.CombinedOutput()
			if ctx.Err() == context.DeadlineExceeded {
				log.Error(err, "Command timed out")
				return err
			}
			if err != nil {
				log.Error(err, "Command failed", "output", string(out))
			}
			return err
		}
	}

	return nil
}

func GetGitRootDir(spec *opsv1alpha1.GitOps) string {
	return spec.Status.RootFolder + "/" + spec.Spec.RootFolder
}
