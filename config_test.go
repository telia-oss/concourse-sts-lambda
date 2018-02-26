package main_test

import (
	"encoding/json"
	pkg "github.com/TeliaSoneraNorge/concourse-sts-lambda"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	input := strings.TrimSpace(`
{
    "name": "team",
    "keyId": "key",
    "accounts": [
	{
	    "name": "account",
	    "roleArn": "role"
	}
    ]
}
`)

	t.Run("Unmarshal works as intended", func(t *testing.T) {
		expected := pkg.Team{
			Name:  "team",
			KeyID: "key",
			Accounts: []*pkg.Account{
				{
					Name:    "account",
					RoleArn: "role",
				},
			},
		}

		var actual pkg.Team
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
		actual, err := pkg.NewPath(team, account, template).String()
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Fails if template expects additional parameters", func(t *testing.T) {
		template := "/concourse/{{.Team}}/{{.Account}}/{{.Something}}"
		actual, err := pkg.NewPath(team, account, template).String()
		assert.NotNil(t, err)
		assert.Equal(t, "", actual)
	})
}
