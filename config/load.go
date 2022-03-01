package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/flosch/pongo2"
	"github.com/onemoreteam/yaml"
)

func BytesFromFile(fp string) (_ []byte, err error) {
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
		return nil, fmt.Errorf("unknown config file extension: %v", ext)
	}
	return b, nil
}

func FromFile(fp string, cfg interface{}) (err error) {
	b, err := BytesFromFile(fp)
	if err != nil {
		return
	}
	return FromJson(b, cfg)
}

func FromJson(b []byte, cfg interface{}) (err error) {
	return json.Unmarshal(b, cfg)
}

func FromJsonString(s string, cfg interface{}) (err error) {
	return
}
