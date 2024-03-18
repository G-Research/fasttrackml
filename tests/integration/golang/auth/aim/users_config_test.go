package aim

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/zeebo/assert"
	"gopkg.in/yaml.v3"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config/auth"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UsersConfigAuthTestSuite struct {
	helpers.BaseTestSuite
}

func TestUsersConfigAuthTestSuite(t *testing.T) {
	// create users configuration firstly.
	data, err := yaml.Marshal(auth.YamlConfig{
		Users: []auth.YamlUserConfig{
			{
				Name: "user1",
				Roles: []string{
					"ns:namespace1",
					"ns:namespace2",
				},
				Password: "user1password",
			},
			{
				Name: "user2",
				Roles: []string{
					"ns:namespace2",
					"ns:namespace3",
				},
				Password: "user2password",
			},
		},
	})
	assert.Nil(t, err)

	configPath := fmt.Sprintf("%s/users-config.yaml", t.TempDir())
	f, err := os.Create(configPath)
	_, err = f.Write(data)
	assert.Nil(t, err)
	assert.Nil(t, f.Close())

	// run test suite with newly created configuration.
	testSuite := new(UsersConfigAuthTestSuite)
	testSuite.Config = config.ServiceConfig{
		Auth: auth.Config{
			AuthUsersConfig: configPath,
		},
	}
	suite.Run(t, testSuite)
}

func (s *UsersConfigAuthTestSuite) Test_Ok() {
	// create test namespaces.
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  2,
		Code:                "namespace1",
		Description:         "Test namespace 1",
		DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
	})
	s.Require().Nil(err)
	_, err = s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  3,
		Code:                "namespace2",
		Description:         "Test namespace 2",
		DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
	})
	s.Require().Nil(err)
	_, err = s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  4,
		Code:                "namespace3",
		Description:         "Test namespace 3",
		DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
	})
	s.Require().Nil(err)
	_, err = s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  5,
		Code:                "namespace4",
		Description:         "Test namespace 4",
		DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
	})
	s.Require().Nil(err)
}
