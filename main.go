package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

var config Config

// Config for the Lambda function (set via environment).
type Config struct {
	Region string `long:"region" env:"REGION" default:"eu-west-1" description:"AWS region to use for API calls."`
	Bucket string `long:"bucket" env:"CONFIG_BUCKET" default:"" description:"Config bucket name."`
	Key    string `long:"key" env:"CONFIG_KEY" default:"" description:"Config bucket key."`
}

// Team represents a unit which requires dynamic credentials for one or more accounts.
type Team struct {
	Name     string     `json:"name"`
	KeyID    string     `json:"keyId"`
	Accounts []*Account `json:"accounts"`
}

// Account represents a role to assume for a specific account.
// AWS Access key id and secret key will be output as:
// <SSMPath>/<Name>-access-key and <Name>-secret-key
type Account struct {
	Name    string `json:"name"`
	RoleArn string `json:"roleArn"`
}

func getS3Config(sess *session.Session, region, bucket, key string) (output []*Team, err error) {
	config := &aws.Config{Region: aws.String(config.Region)}
	client := s3.New(sess, config)

	res, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	r := json.NewDecoder(res.Body)
	if err := r.Decode(&output); err != nil {
		return nil, err
	}

	return output, nil
}

func assumeRole(sess *session.Session, account *Account) (*sts.Credentials, error) {
	config := &aws.Config{Region: aws.String(config.Region)}
	client := sts.New(sess, config)
	input := &sts.AssumeRoleInput{
		DurationSeconds: aws.Int64(3600),
		RoleArn:         aws.String(account.RoleArn),
		RoleSessionName: aws.String(account.Name),
	}
	output, err := client.AssumeRole(input)
	if err != nil {
		return nil, err
	}
	return output.Credentials, nil
}

func writeCredentials(sess *session.Session, team *Team, account *Account, credentials *sts.Credentials) error {
	config := &aws.Config{Region: aws.String(config.Region)}
	client := ssm.New(sess, config)

	// Access key
	err := writeSecret(
		client,
		aws.String(team.KeyID),
		aws.String(fmt.Sprintf("/concourse/%s/%s-access-key", team.Name, account.Name)),
		credentials.AccessKeyId,
	)
	if err != nil {
		return err
	}

	// Secret key
	err = writeSecret(
		client,
		aws.String(team.KeyID),
		aws.String(fmt.Sprintf("/concourse/%s/%s-secret-key", team.Name, account.Name)),
		credentials.SecretAccessKey,
	)
	if err != nil {
		return err
	}

	// Token
	err = writeSecret(
		client,
		aws.String(team.KeyID),
		aws.String(fmt.Sprintf("/concourse/%s/%s-session-token", team.Name, account.Name)),
		credentials.SessionToken,
	)
	if err != nil {
		return err
	}

	// Expiration
	err = writeSecret(
		client,
		aws.String(team.KeyID),
		aws.String(fmt.Sprintf("/concourse/%s/%s-expiration", team.Name, account.Name)),
		aws.String(credentials.Expiration.Format("2006-01-02 15:04")),
	)
	return err
}

func writeSecret(client *ssm.SSM, key, name, value *string) error {
	input := &ssm.PutParameterInput{
		Name:      name,
		Value:     value,
		Type:      aws.String("SecureString"),
		KeyId:     key,
		Overwrite: aws.Bool(true),
	}
	_, err := client.PutParameter(input)
	return err
}

// Handler for Lambda
func Handler() error {
	// Parse and validate Config
	_, err := flags.Parse(&config)
	if err != nil {
		return err
	}
	if config.Bucket == "" {
		return errors.New("missing CONFIG_BUCKET")
	}
	if config.Key == "" {
		return errors.New("missing CONFIG_KEY")
	}

	// Create a new session
	sess, err := session.NewSession()
	if err != nil {
		return errors.Wrap(err, "failed to create new session")
	}

	// Get config from S3
	teams, err := getS3Config(sess, config.Region, config.Bucket, config.Key)
	if err != nil {
		return errors.Wrap(err, "failed to get config from s3")
	}

	// Loop through teams and assume roles/write credentials for
	// all accounts controlled by the team.
	for _, team := range teams {
		for _, account := range team.Accounts {
			creds, err := assumeRole(sess, account)
			if err != nil {
				return errors.Wrapf(err, "failed to assume role: %s", account.RoleArn)
			}
			err = writeCredentials(sess, team, account, creds)
			if err != nil {
				return errors.Wrap(err, "failed to write credentials: %s")
			}
		}
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
