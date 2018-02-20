package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

// Command options
type Command struct {
	Region string `long:"region" env:"REGION" default:"eu-west-1" description:"AWS region to use for API calls."`
	Bucket string `long:"bucket" env:"CONFIG_BUCKET" default:"" description:"Config bucket name."`
	Key    string `long:"key" env:"CONFIG_KEY" default:"" description:"Config bucket key."`
}

// Validate the Command options.
func (c *Command) Validate() error {
	if c.Region == "" {
		return errors.New("missing CONFIG_REGION")
	}
	if c.Bucket == "" {
		return errors.New("missing CONFIG_BUCKET")
	}
	if c.Key == "" {
		return errors.New("missing CONFIG_KEY")
	}
	return nil
}

// Handler for Lambda
func Handler() error {
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

	// Load config from S3
	config, err := manager.ReadConfig(command.Bucket, command.Key)
	if err != nil {
		return errors.Wrap(err, "failed to get config from s3")
	}

	// Loop through teams and assume roles/write credentials for
	// all accounts controlled by the team.
	for _, team := range *config {
		for _, account := range team.Accounts {
			creds, err := manager.AssumeRole(account.RoleArn, team.Name)
			if err != nil {
				return errors.Wrapf(err, "failed to assume role: %s", account.RoleArn)
			}

			path := fmt.Sprintf("/concourse/%s/%s", team.Name, account.Name)
			if err := manager.WriteCredentials(creds, path, team.KeyID); err != nil {
				return errors.Wrap(err, "failed to write credentials: %s")
			}
		}
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
