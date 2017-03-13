package descriptor_test

import (
	"testing"

	"github.com/ThingiverseIO/thingiverseio/descriptor"
)

var testdesc = `
func funcname(param1 string, param2 []int) (outp1 string, outp2 []bool)
func fun1(param1 string, param2 []int) ()
func fun2() (outp1 string, outp2 []bool)
tag simple_tag
tag key_tag: tag_value
#define multiple tags in one line
tags multisimple, multikey:val
`

func TestParseDescriptor(t *testing.T) {
	desc, err := descriptor.Parse(testdesc)
	if err != nil {
		t.Error(err)
	}
	if len(desc.Functions) != 3 {
		t.Error("Wrong number of functions, want 3, got", len(desc.Functions))
	}
	if len(desc.Functions[0].Input) != 2 {
		t.Error("Wrong number of input parameters, want 2, got", len(desc.Functions[0].Input))
	}
	if len(desc.Functions[0].Output) != 2 {
		t.Error("Wrong number of output parameters, want 2, got", len(desc.Functions[0].Output))
	}

	if len(desc.Tags) != 4 {
		t.Error("Wrong Number of tags, want 4, got", len(desc.Tags))
	}
}
