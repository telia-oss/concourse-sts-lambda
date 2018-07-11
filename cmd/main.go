package main

import (
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"github.com/telia-oss/concourse-sts-lambda"
)

// Command options
type Command struct {
	Region   string `long:"region" env:"REGION" description:"AWS region to use for API calls."`
	Path     string `long:"secrets-manager-path" env:"SECRETS_MANAGER_PATH" default:"/concourse/{{.Team}}/{{.Account}}" description:"Path to use when writing to AWS Secrets manager."`
	KmsKeyID string `long:"kms-key-id" env:"KMS_KEY_ID" default:"" description:"KMS Key ID (or ALIAS/ARN) used to encrypt the secrets."`
}

// Validate the Command options.
func (c *Command) Validate() error {
	if c.Region == "" {
		return errors.New("missing required argument 'region'")
	}
	return nil
}

func main() {
	var command Command

	_, err := flags.Parse(&command)
	if err != nil {
		panic(fmt.Errorf("failed to parse flag %s", err))
	}
	if err := command.Validate(); err != nil {
		panic(fmt.Errorf("invalid command: %s", err))
	}
	sess, err := session.NewSession()
	if err != nil {
		panic(fmt.Errorf("failed to create new session: %s", err))
	}

	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}

	f := handler.New(handler.NewManager(sess, command.Region), command.Path, command.KmsKeyID, logger)
	lambda.Start(f)
}
