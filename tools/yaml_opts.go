package tools

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

func ReadYamlFileInPath(configPathsStr string, out interface{}) error {
	configRead := false
	configPaths := strings.Split(configPathsStr, ",")
	for _, configPath := range configPaths {
		if !FileExists(configPath) {
			continue
		}
		if err := ReadYamlFile(configPath, out); err != nil {
			return fmt.Errorf("unable to read config in %s: %s", configPath, err)
		}
		configRead = true
	}
	if !configRead {
		return fmt.Errorf("no config files found in %v", configPaths)
	}
	return nil
}

func ReadYamlFile(path string, out interface{}) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if b == nil || len(b) < 1 {
		return errors.New("missing configuration data")
	}
	if err := yaml.Unmarshal(b, out); err != nil {
		return err
	}
	return nil
}
