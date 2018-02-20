package main_test

import (
	"encoding/json"
	pkg "github.com/itsdalmo/sts-lambda"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	input := strings.TrimSpace(`
[
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
]
`)

	t.Run("Unmarshal works as intended", func(t *testing.T) {
		expected := pkg.Config{
			&pkg.Team{
				Name:  "team",
				KeyID: "key",
				Accounts: []*pkg.Account{
					{
						Name:    "account",
						RoleArn: "role",
					},
				},
			},
		}

		var actual pkg.Config
		err := json.Unmarshal([]byte(input), &actual)

		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})
}
