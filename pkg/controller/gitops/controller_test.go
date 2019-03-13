package gitops

import (
	"testing"
	"time"

	"github.com/nokia/gitops-conductor/pkg/apis/ops/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
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

func TestFilterObject(t *testing.T) {

	r := ReconcileGitOps{rensureInterval: 20}
	servicAcccount := &corev1.ServiceAccount{}
	service := &corev1.Service{}
	assert.True(t, r.filterObject(servicAcccount))
	assert.False(t, r.filterObject(service))

}
