package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ThingiverseIO/thingiverseio/descriptor"

	"gopkg.in/gcfg.v1"
)

type CfgFile struct {
	Network  CfgFileNetwork
	Misc     CfgFileMisc
	UserTags CfgFileUserTags
}

type CfgFileNetwork struct {
	Interface []string
}

type CfgFileMisc struct {
	Logging string
	Debug   bool
}

type CfgFileUserTags struct {
	Tag []string
}

func CheckCfgFile(cfg *UserConfig, path string) {

	cfgf, err := ReadCfgFile(path)

	if err != nil {
		return
	}

	if len(cfgf.Network.Interface) != 0 {
		cfg.Interfaces = cfgf.Network.Interface
	}

	setLoggerFromString(cfgf.Misc.Logging, cfg)
	parseUserTags(cfgf.UserTags.Tag, cfg)
}

func ReadCfgFile(path string) (cfgf CfgFile, err error) {

	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	err = gcfg.ReadInto(&cfgf, f)

	return

}

func WriteCfgFile(cfgf CfgFile, path string) (err error) {

	var lines []string

	if len(cfgf.Network.Interface) != 0 {

		lines = append(lines, "[usertags]")
		for _, iface := range cfgf.Network.Interface {
			lines = append(lines, fmt.Sprintf("interface=%s", iface))
		}

		lines = append(lines, "")
	}

	lines = append(lines, "[misc]")

	if cfgf.Misc.Logging != "" {
		lines = append(lines, fmt.Sprintf("logging=%s", cfgf.Misc.Logging))
	}

	debug := "false"
	if cfgf.Misc.Debug {
		debug = "true"
	}
	lines = append(lines, fmt.Sprintf("debug=%s", debug))

	lines = append(lines, "[usertags]")

	for _, tag := range cfgf.UserTags.Tag {
		lines = append(lines, fmt.Sprintf("tag=%s", tag))
	}

	if _, err = os.Stat(path); err == nil {
		err = os.Remove(path)
		if err != nil {
			return
		}
	}

	err = ioutil.WriteFile(path, []byte(strings.Join(lines, "\n")), 0777)

	return
}

func parseUserTags(tags []string, cfg *UserConfig) {
	for _, t := range tags {
		var tag descriptor.Tag
		tag.Scan(t)
		cfg.Tags.Add(tag)
	}
}
