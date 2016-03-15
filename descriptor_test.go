package thingiverseio_test

import (
	"testing"

	"github.com/joernweissenborn/thingiverseio"
)

var testdesc = `
func funcname(param1 string, param2 []int) (outp1 string, outp2 []bool)
tag simple_tag
tag key_tag: tag_value
#define multiple tags in one line
tags multisimple muiltikey:val
`

func TestParseDescriptor(t *testing.T) {
	_, err := thingiverseio.ParseDescriptor(testdesc)
	if err !=nil {
		t.Error(err)
	}
}
