package confd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_confdTomlGen(t *testing.T) {
	confdt := ConfdTomlT{
		Prefix:   "test",
		ConfPath: "basic.conf",
	}
	err := confdt.ConfdTomlGen()
	assert.Nil(t, err, "confdTomlGen fail")
}
