package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"log"
)

// Command options
type Command struct {
	Region string `long:"region" env:"REGION" default:"eu-west-1" description:"AWS region to use for API calls."`
	Path   string `long:"path" env:"SSM_PATH" default:"/concourse/{{.Team}}/{{.Account}}" description:"Path to use when writing to SSM."`
}

// Validate the Command options.
func (c *Command) Validate() error {
	if c.Region == "" {
		return errors.New("missing CONFIG_REGION")
	}
	return nil
}

// Handler for Lambda
func Handler(team Team) error {
	var command Command

	// Parse flags and validate
	_, err := flags.Parse(&command)
	if err != nil {
		return errors.Wrap(err, "failed to parse flags")
	}
	if err := command.Validate(); err != nil {
		return errors.Wrap(err, "invalid command")
	}

	// New session and manager
	sess := session.Must(session.NewSession())
	manager := NewManager(sess, command.Region)

	// Loop through teams and assume roles/write credentials for
	// all accounts controlled by the team.
	for _, account := range team.Accounts {
		creds, err := manager.AssumeRole(account.RoleArn, team.Name)
		if err != nil {
			log.Printf("failed to assume role (%s): %s", account.RoleArn, err)
			continue
		}

		path, err := NewPath(team.Name, account.Name, command.Path).String()
		if err != nil {
			log.Printf("failed to parse secret path: %s", err)
			continue
		}
		if err := manager.WriteCredentials(creds, path, team.KeyID); err != nil {
			log.Printf("failed to write credentials: %s", err)
			continue
		}
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
