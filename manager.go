package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

// SecretsClient for testing purposes.
//go:generate mockgen -destination=mocks/mock_secrets_client.go -package=mocks github.com/telia-oss/concourse-sts-lambda SecretsClient
type SecretsClient secretsmanageriface.SecretsManagerAPI

// STSClient for testing purposes.
//go:generate mockgen -destination=mocks/mock_sts_client.go -package=mocks github.com/telia-oss/concourse-sts-lambda STSClient
type STSClient stsiface.STSAPI

// S3Client for testing purposes.
//go:generate mockgen -destination=mocks/mock_s3_client.go -package=mocks github.com/telia-oss/concourse-sts-lambda S3Client
type S3Client s3iface.S3API

// Manager handles API calls to AWS.
type Manager struct {
	secretsClient SecretsClient
	stsClient     STSClient
	s3Client      S3Client
}

// NewManager creates a new manager from an existing AWS session.
func NewManager(sess *session.Session) *Manager {
	return &Manager{
		stsClient:     sts.New(sess),
		secretsClient: secretsmanager.New(sess),
		s3Client:      s3.New(sess),
	}
}

// NewTestManager ...
func NewTestManager(sm SecretsClient, sts STSClient, s3 S3Client) *Manager {
	return &Manager{secretsClient: sm, stsClient: sts, s3Client: s3}
}

// ReadConfig from S3.
func (m *Manager) ReadConfig(bucket, key string) (*Team, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	obj, err := m.s3Client.GetObject(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object from s3: %s", err)
	}
	defer obj.Body.Close()

	b := bytes.NewBuffer(nil)
	if _, err := io.Copy(b, obj.Body); err != nil {
		return nil, fmt.Errorf("failed to copy object body: %s", err)
	}

	var out *Team
	if err := json.Unmarshal(b.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal object: %s", err)
	}
	return out, nil
}

// AssumeRole on the given role ARN and the given team name (identifier).
func (m *Manager) AssumeRole(arn, team string, duration int64) (*sts.Credentials, error) {
	if duration == 0 {
		duration = 3600
	}
	input := &sts.AssumeRoleInput{
		DurationSeconds: aws.Int64(duration),
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
func (m *Manager) WriteCredentials(creds *sts.Credentials, path string) error {
	values := map[string]string{
		path + "-access-key":    aws.StringValue(creds.AccessKeyId),
		path + "-secret-key":    aws.StringValue(creds.SecretAccessKey),
		path + "-session-token": aws.StringValue(creds.SessionToken),
	}

	for name, value := range values {
		err := m.writeSecret(name, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) writeSecret(name, secret string) error {
	var err error
	// Fewer API calls to naively try to create it and handle the error.
	_, err = m.secretsClient.CreateSecret(&secretsmanager.CreateSecretInput{
		Name:        aws.String(name),
		Description: aws.String("STS Credentials for Concourse."),
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

	timestamp := time.Now().Format(time.RFC3339)

	_, err = m.secretsClient.UpdateSecret(&secretsmanager.UpdateSecretInput{
		Description:  aws.String(fmt.Sprintf("STS Credentials for Concourse. Last updated: %s", timestamp)),
		SecretId:     aws.String(name),
		SecretString: aws.String(secret),
	})
	return err
}
