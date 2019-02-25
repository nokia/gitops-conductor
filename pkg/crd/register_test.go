package crd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadCrdFile(t *testing.T) {
	c, err := readConfig("config.yaml")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(c.Crds))
}
