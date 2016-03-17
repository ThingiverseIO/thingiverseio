package thingiverseio

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
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

func ParseDescriptor(desc string) (d Descriptor, err error) {
	scanner := bufio.NewScanner(strings.NewReader(desc))
	linecounter := 0
	d.Tags = map[string]string{}
	for scanner.Scan() {
		line := strings.TrimLeft(scanner.Text(), " ")
		linecounter++
		switch {
		case line == "":
		case strings.HasPrefix(line, "#"):
		case strings.HasPrefix(line, "func"):
			var f Function
			f, err = parseFunction(linecounter, line)
			if err != nil {
				return
			}
			d.Functions = append(d.Functions, f)
		case strings.HasPrefix(line, "tags"):
			var tags map[string]string
			tags, err = parseTagsLine(linecounter, line)
			if err != nil {
				return
			}
			for k, v := range tags {
				d.Tags[k] = v
			}
		case strings.HasPrefix(line, "tag"):
			var k, v string
			k, v, err = parseTagLine(linecounter, line)
			if err != nil {
				return
			}
			d.Tags[k] = v
		default:
			err = newLineError(linecounter, "malformed line")
			return
		}
	}
	return
}

func parseFunction(line int, s string) (f Function, err error) {
	s = strings.TrimLeft(s, "func")
	s = strings.TrimLeft(s, " ")
	split1 := strings.Split(s, "(")
	f.Name = strings.TrimRight(split1[0], " ")

	if len(split1) == 1 {
		return
	}

	ins := strings.TrimRight(strings.TrimRight(split1[1], " "), ")")
	split2 := strings.Split(ins, ",")

	for _, in := range split2 {
		in = strings.TrimLeft(in, " ")
		in = strings.TrimRight(in, " ")
		if len(in) == 0 {
			break
		}
		spl := strings.Split(in, " ")
		if len(spl) != 2 {
			err = newLineError(line, fmt.Sprint("malformed function input parameter", in))
			return
		}
		n := spl[0]
		t := spl[1]

		if !containsAny(t, "string", "bool", "bin", "int", "float") {
			err = newLineError(line, fmt.Sprint("malformed function, unknown type", t))
			return
		}
		f.Input = append(f.Input, Parameter{n, t})
	}

	if len(split1) == 2 {
		return
	}

	outs := strings.TrimRight(strings.TrimRight(split1[2], " "), ")")
	split2 = strings.Split(outs, ",")
	for _, out := range split2 {
		out = strings.TrimLeft(out, " ")
		out = strings.TrimRight(out, " ")
		if len(out) == 0 {
			break
		}
		spl := strings.Split(out, " ")
		if len(spl) != 2 {
			err = newLineError(line, fmt.Sprint("malformed function output parameter", out))
			return
		}
		n := spl[0]
		t := spl[1]

		if !containsAny(t, "string", "bool", "bin", "int", "float") {
			err = newLineError(line, fmt.Sprint("malformed function, unknown type", t))
			return
		}
		f.Output = append(f.Input, Parameter{n, t})
	}

	return
}

func parseTagLine(line int, s string) (k, v string, err error) {
	s = strings.TrimLeft(s, "tag")
	return parseTag(line, s)
}

func parseTagsLine(line int, s string) (tags map[string]string, err error) {
	tags = map[string]string{}
	s = strings.TrimLeft(s, "tags")
	split := strings.Split(s, ",")
	for _, t := range split {
		var k, v string
		k, v, err = parseTag(line, t)
		if err != nil {
			return
		}
		tags[k] = v
	}
	return
}

func parseTag(line int, s string) (k, v string, err error) {
	if strings.Contains(s, ",") {
		err = newLineError(line, "malformed tag, ',' not allowed in tag")
		return
	}
	if strings.Contains(s, ":") {
		split := strings.Split(s, ":")
		if len(split) != 2 {
			err = newLineError(line, "malformed tag")
			return
		}
		k = strings.Trim(split[0], " ")
		v = strings.Trim(split[1], " ")
		return
	} else {
		k = strings.Trim(s, " ")
	}
	return
}

func newLineError(line int, reason string) error {
	return errors.New(fmt.Sprintf("LINE %d: %s", line, reason))
}

func containsAny(s string, t ...string) bool {
	for _, str := range t {
		if strings.Contains(s, str) {
			return true
		}
	}
	return false
}
