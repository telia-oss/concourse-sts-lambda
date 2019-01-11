package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	handler "github.com/telia-oss/concourse-sts-lambda"
)

// Command options
type Command struct {
	Path string `envconfig:"SECRETS_MANAGER_PATH" default:"/concourse/{{.Team}}/{{.Account}}"`
}

func main() {
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}

	var command Command
	err := envconfig.Process("", command)
	if err != nil {
		logger.Fatalf("failed to parse envconfig: %s", err)
	}
	sess, err := session.NewSession()
	if err != nil {
		logger.Fatalf("failed to create a new session: %s", err)
	}

	f := handler.New(handler.NewManager(sess), command.Path, logger)
	lambda.Start(f)
}
