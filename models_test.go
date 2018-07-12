package handler_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/telia-oss/concourse-sts-lambda"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		description string
		input       string
		expected    handler.Team
	}{
		{
			description: "Unmarshal works as intended",
			input: strings.TrimSpace(`
{
    "name": "team",
    "accounts": [
	{
	    "name": "account",
	    "roleArn": "role"
	}
    ]
}
`),
			expected: handler.Team{
				Name: "team",
				Accounts: []*handler.Account{
					{
						Name:    "account",
						RoleArn: "role",
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			var output handler.Team
			err := json.Unmarshal([]byte(tc.input), &output)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if got, want := output, tc.expected; !reflect.DeepEqual(got, want) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, want)
			}
		})
	}
}

func TestSecretPath(t *testing.T) {
	tests := []struct {
		description string
		template    string
		team        string
		account     string
		expected    string
		shouldError bool
	}{
		{
			description: "template works as intended",
			template:    "/concourse/{{.Team}}/{{.Account}}",
			team:        "TEAM",
			account:     "ACCOUNT",
			expected:    "/concourse/TEAM/ACCOUNT",
			shouldError: false,
		},
		{
			description: "fails if the template expects more parameters",
			template:    "/concourse/{{.Team}}/{{.Account}}/{{.Something}}",
			team:        "TEAM",
			account:     "ACCOUNT",
			expected:    "",
			shouldError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			got, err := handler.NewSecretPath(tc.team, tc.account, tc.template).String()

			if tc.shouldError && err == nil {
				t.Fatal("expected an error to occur")
			}

			if !tc.shouldError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if want := tc.expected; got != want {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, want)
			}
		})
	}
}
