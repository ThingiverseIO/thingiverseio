package thingiverseio

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

type Descriptor struct {
	Functions []Function
	Tags      map[string]string
}

type Function struct {
	Name   string
	Input  []Parameter
	Output []Parameter
}

func (f Function) String() string {
	inpar := ""
	for _, p := range f.Input {
		inpar = fmt.Sprintf("%s%s", inpar, p)
	}
	outpar := ""
	for _, p := range f.Output {
		outpar = fmt.Sprintf("%s%s", outpar, p)
	}
	return fmt.Sprintf("%s%s%s", f.Name, inpar, outpar)
}

type Parameter struct {
	Name string
	Type string
}

//name(par:type,...)par:type

func (p Parameter) String() string {
	return fmt.Sprintf("%s%s", p.Name, p.Type)
}

func (a Descriptor) AsTagSet() (tagset map[string]string) {

	if a.Tags != nil {
		tagset = a.Tags
	} else {
		tagset = map[string]string{}
	}
	for _, fn := range a.Functions {
		tagset[fmt.Sprintf("%s", fn)] = "f"
	}
	return
}

func FromJson(JSON string) (dsc *Descriptor) {
	dsc = &Descriptor{}
	err := json.Unmarshal([]byte(JSON), dsc)
	if err != nil {
		panic(fmt.Sprint("Insane ServiceDescriptor", err))
	}
	return
}

func FromYaml(YAML string) (dsc *Descriptor) {
	dsc = &Descriptor{}
	if err := yaml.Unmarshal([]byte(YAML), dsc); err != nil {
		panic(fmt.Sprint("Insane ServiceDescriptor", err))
	}
	return
}
