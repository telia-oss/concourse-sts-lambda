package handler_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/telia-oss/concourse-sts-lambda"
)

func TestConfig(t *testing.T) {
	input := strings.TrimSpace(`
{
    "name": "team",
    "accounts": [
	{
	    "name": "account",
	    "roleArn": "role"
	}
    ]
}
`)

	t.Run("Unmarshal works as intended", func(t *testing.T) {
		expected := handler.Team{
			Name: "team",
			Accounts: []*handler.Account{
				{
					Name:    "account",
					RoleArn: "role",
				},
			},
		}

		var actual handler.Team
		err := json.Unmarshal([]byte(input), &actual)

		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestSecretPath(t *testing.T) {
	var (
		team    = "TEAM"
		account = "ACCOUNT"
	)

	t.Run("Secret template works as intended", func(t *testing.T) {
		template := "/concourse/{{.Team}}/{{.Account}}"
		expected := "/concourse/TEAM/ACCOUNT"
		actual, err := handler.NewSecretPath(team, account, template).String()
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Fails if template expects additional parameters", func(t *testing.T) {
		template := "/concourse/{{.Team}}/{{.Account}}/{{.Something}}"
		actual, err := handler.NewSecretPath(team, account, template).String()
		assert.NotNil(t, err)
		assert.Equal(t, "", actual)
	})
}
