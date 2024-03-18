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

// UserPermissions represents permission object into which the user configuration is parsed.
type UserPermissions struct {
	data map[string]map[string]struct{}
}

// HasAccess makes check that user has permission to access to the requested namespace.
func (p UserPermissions) HasAccess(namespace string, authToken string) bool {
	if authToken == "" {
		return false
	}

	roles, ok := p.data[authToken]
	if !ok {
		return ok
	}

	if _, ok := roles["admin"]; ok {
		return true
	}

	if _, ok := roles[fmt.Sprintf("ns:%s", namespace)]; !ok {
		return ok
	}
	return true
}

// Load loads user configuration from given configuration file.
func Load(configFilePath string) (*UserPermissions, error) {
	//nolint:gosec
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, eris.Wrap(err, "error reading user configuration file")
	}

	switch filepath.Ext(configFilePath) {
	case ".yaml", ".yml":
		permissions, err := parseUserConfigFromYaml(data)
		if err != nil {
			return nil, eris.Wrap(err, "error parsing user configuration from yaml")
		}
		return permissions, nil
	}
	return nil, eris.Errorf("unsupported user configuration file type")
}

// YamlConfig represents users configuration in YAML format.
type YamlConfig struct {
	Users []YamlUserConfig `yaml:"users"`
}

// YamlUserConfig partial object of YamlConfig.
type YamlUserConfig struct {
	Name     string   `yaml:"name"`
	Password string   `yaml:"password"`
	Roles    []string `yaml:"roles"`
}

// parseUserConfigFromYaml parse configuration from ".yaml", ".yml" files and transform it into internal representation.
func parseUserConfigFromYaml(content []byte) (*UserPermissions, error) {
	config := YamlConfig{}
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, eris.Wrap(err, "error unmarshaling data from yaml file")
	}

	permissions := UserPermissions{data: make(map[string]map[string]struct{})}
	passwordRegex := regexp.MustCompile(`^\$\{(.*)\}$`)
	passwordReplacer := strings.NewReplacer("$", "", "{", "", "}", "")
	for _, user := range config.Users {
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
		permissions.data[loginEncoded] = roles
	}

	return &permissions, nil
}
