package gitops

import (
	"testing"
	"time"

	"github.com/nokia/gitops-conductor/pkg/apis/ops/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestCalcTime(t *testing.T) {

	r := ReconcileGitOps{rensureInterval: 20}
	instance := v1alpha1.GitOps{}
	instance.Status.Updated = "10:57:01"
	nowTime, _ := time.Parse("15:04:05", "11:55:01")
	isover := r.isOverDuration(nowTime, &instance)
	assert.Equal(t, true, isover)

}

func TestCalcNotOver(t *testing.T) {

	r := ReconcileGitOps{rensureInterval: 20}
	instance := v1alpha1.GitOps{}
	instance.Status.Updated = "10:57:01"

	nowTime, _ := time.Parse("15:04:05", "10:59:01")
	isover := r.isOverDuration(nowTime, &instance)
	assert.False(t, isover)
}
