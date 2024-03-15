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

// Permissions represents permission object into which the RBAC configuration is parsed.
type Permissions struct {
	data map[string]map[string]struct{}
}

// HasPermissions makes check that user has permission to access to the requested namespace.
func (p Permissions) HasPermissions(namespace string, authToken string) bool {
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

// Load loads RBAC configuration from given configuration file.
func Load(configFilePath string) (*Permissions, error) {
	//nolint:gosec
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, eris.Wrap(err, "error reading rbac configuration file")
	}

	switch filepath.Ext(configFilePath) {
	case ".yaml", ".yml":
		permissions, err := parsePermissionFromYaml(data)
		if err != nil {
			return nil, eris.Wrap(err, "error parsing rbac configuration from yaml")
		}
		return permissions, nil
	}
	return nil, eris.Errorf("unsupported rbac configuration file type")
}

// parsePermissionFromYaml parse configuration from ".yaml", ".yml" files and transform it into internal representation.
func parsePermissionFromYaml(content []byte) (*Permissions, error) {
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

	permissions := Permissions{data: make(map[string]map[string]struct{})}
	passwordRegex := regexp.MustCompile(`^\$\{(.*)\}$`)
	passwordReplacer := strings.NewReplacer("$", "", "{", "", "}", "")
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
		permissions.data[loginEncoded] = roles
	}

	return &permissions, nil
}
