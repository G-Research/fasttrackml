package auth

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rotisserie/eris"
	"gopkg.in/yaml.v3"
)

// Load loads RBAC configuration from given configuration file.
func Load(configFilePath string) (map[string]map[string]struct{}, error) {
	//nolint:gosec
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, eris.Wrap(err, "error reading rbac configuration file")
	}

	switch filepath.Ext(configFilePath) {
	case ".yaml", ".yml":
		config, err := parseYamlConfiguration(data)
		if err != nil {
			return nil, eris.Wrap(err, "error parsing rbac configuration from yaml")
		}
		return config, nil
	}
	return nil, eris.Errorf("unsupported rbac configuration file type")
}

// parseYamlConfiguration parse configuration from ".yaml", ".yml" files and transform it into internal representation.
func parseYamlConfiguration(content []byte) (map[string]map[string]struct{}, error) {
	type config struct {
		Users []struct {
			Name     string   `yaml:"name"`
			Password string   `yaml:"password"`
			Roles    []string `yaml:"roles"`
		} `yaml:"users"`
	}

	cfg := config{}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, eris.Wrap(err, "error unmarshaling data from yaml file")
	}

	data := map[string]map[string]struct{}{}
	passwordRegex, passwordReplacer := regexp.MustCompile(`^\$\{(.*)\}$`), strings.NewReplacer("$", "", "{", "", "}", "")
	for _, user := range cfg.Users {
		// if password format is ${PASSWORD_PARAMETER_FROM_ENV} then try to load it from ENV.
		if passwordRegex.MatchString(user.Password) {
			password, ok := os.LookupEnv(passwordReplacer.Replace(user.Password))
			if !ok {
				return nil, eris.Errorf("error reading password from ENV variable: %s", user.Password)
			}
			user.Password = password
		}
		roles := map[string]struct{}{}
		for _, role := range user.Roles {
			roles[role] = struct{}{}
		}

		// encode name + password into base64. it helps later to quickly access/find user,
		// so we won't have any performance degradation.
		loginEncoded := base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", user.Name, user.Password)),
		)
		data[loginEncoded] = roles
	}

	return data, nil
}
