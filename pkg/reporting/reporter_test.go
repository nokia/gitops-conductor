package reporting

import (
	"io/ioutil"
	"testing"

	"github.com/nokia/gitops-conductor/tests/utils"
	"github.com/stretchr/testify/assert"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func TestCollectFile(t *testing.T) {
	logf.SetLogger(&utils.Logg{})
	data := `tags:
  - key: test
    value: good`

	err := ioutil.WriteFile("/tmp/update_result.yaml", []byte(data), 0644)
	assert.Nil(t, err)
	tags := collectTagFile()
	assert.Len(t, tags, 1, "Invalid length on tags")
}
