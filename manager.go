package handler

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

// SecretsManager for testing purposes.
//go:generate mockgen -destination=mocks/mock_secretsmanager.go -package=mocks github.com/telia-oss/concourse-sts-lambda SecretsManager
type SecretsManager secretsmanageriface.SecretsManagerAPI

// STSManager for testing purposes.
//go:generate mockgen -destination=mocks/mock_stsmanager.go -package=mocks github.com/telia-oss/concourse-sts-lambda STSManager
type STSManager stsiface.STSAPI

// Manager handles API calls to AWS.
type Manager struct {
	secretsClient SecretsManager
	stsClient     STSManager
}

// NewManager creates a new manager from a session and region string.
func NewManager(sess *session.Session, region string) *Manager {
	config := &aws.Config{Region: aws.String(region)}
	return &Manager{
		stsClient:     sts.New(sess, config),
		secretsClient: secretsmanager.New(sess, config),
	}
}

// NewTestManager ...
func NewTestManager(s SecretsManager, t STSManager) *Manager {
	return &Manager{secretsClient: s, stsClient: t}
}

// AssumeRole on the given role ARN and the given team name (identifier).
func (m *Manager) AssumeRole(arn, team string) (*sts.Credentials, error) {
	input := &sts.AssumeRoleInput{
		DurationSeconds: aws.Int64(3600),
		RoleArn:         aws.String(arn),
		RoleSessionName: aws.String(team),
	}

	out, err := m.stsClient.AssumeRole(input)
	if err != nil {
		return nil, err
	}
	return out.Credentials, nil
}

// WriteCredentials handles writing a set of Credentials to the parameter store.
func (m *Manager) WriteCredentials(creds *sts.Credentials, path, keyID string) error {
	values := map[string]string{
		path + "-access-key":    aws.StringValue(creds.AccessKeyId),
		path + "-secret-key":    aws.StringValue(creds.SecretAccessKey),
		path + "-session-token": aws.StringValue(creds.SessionToken),
	}

	for name, value := range values {
		err := m.writeSecret(name, value, keyID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) writeSecret(name, secret, keyID string) error {
	var err error

	// Fewer API calls to naively try to create it and handle the error.
	_, err = m.secretsClient.CreateSecret(&secretsmanager.CreateSecretInput{
		Name:        aws.String(name),
		Description: aws.String("Lambda created secret for Concourse."),
		KmsKeyId:    aws.String(keyID),
	})
	if err != nil {
		e, ok := err.(awserr.Error)
		if !ok {
			return fmt.Errorf("failed to convert error: %s", err)
		}
		if e.Code() != secretsmanager.ErrCodeResourceExistsException {
			return err
		}
	}

	_, err = m.secretsClient.PutSecretValue(&secretsmanager.PutSecretValueInput{
		SecretId:      aws.String(name),
		SecretString:  aws.String(secret),
		VersionStages: []*string{aws.String("AWSCURRENT")},
	})
	return err
}
