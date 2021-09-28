package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/flosch/pongo2"
	"github.com/ghodss/yaml"
)

func FromFile(fp string, cfg interface{}) (err error) {
	tpl, err := pongo2.FromFile(fp)
	if err != nil {
		return
	}
	b, err := tpl.ExecuteBytes(nil)
	if err != nil {
		return
	}
	switch ext := filepath.Ext(fp); ext {
	case ".json":
	case ".yml", ".yaml":
		if b, err = yaml.YAMLToJSON(b); err != nil {
			return
		}
	default:
		return fmt.Errorf("unknown config file extension: %v", ext)
	}
	return FromJson(b, cfg)
}

func FromJson(b []byte, cfg interface{}) (err error) {
	return json.Unmarshal(b, cfg)
}

func FromJsonString(s string, cfg interface{}) (err error) {
	return
}
