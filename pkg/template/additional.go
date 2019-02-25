package template

import (
	"io/ioutil"
	"path"

	"text/template"

	"github.com/Masterminds/sprig"
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
	return m

}
