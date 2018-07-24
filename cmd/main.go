package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"github.com/telia-oss/concourse-sts-lambda"
)

// Command options
type Command struct {
	Path string `long:"secrets-manager-path" env:"SECRETS_MANAGER_PATH" default:"/concourse/{{.Team}}/{{.Account}}" description:"Path to use when writing to AWS Secrets manager."`
}

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
}

func main() {
	var command Command

	_, err := flags.Parse(&command)
	if err != nil {
		logger.Fatalf("failed to parse flag: %s", err)
	}
	sess, err := session.NewSession()
	if err != nil {
		logger.Fatalf("failed to create a new session: %s", err)
	}

	f := handler.New(handler.NewManager(sess), command.Path, logger)
	lambda.Start(f)
}
