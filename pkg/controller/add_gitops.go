package controller

import (
	"github.com/nokia/gitops-conductor/pkg/controller/gitops"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, gitops.Add)
}
