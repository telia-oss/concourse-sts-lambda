package main

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

// Manager handles API calls to AWS.
type Manager struct {
	ssmClient ssmiface.SSMAPI
	s3Client  s3iface.S3API
	stsClient stsiface.STSAPI
	region    string
}

// NewManager creates a new manager from a session and region string.
func NewManager(sess *session.Session, region string) *Manager {
	config := &aws.Config{Region: aws.String(region)}
	return &Manager{
		s3Client:  s3.New(sess, config),
		stsClient: sts.New(sess, config),
		ssmClient: ssm.New(sess, config),
		region:    region,
	}
}

// NewTestManager creates a new manager for testing purposes.
func NewTestManager(s3 s3iface.S3API, sts stsiface.STSAPI, ssm ssmiface.SSMAPI) *Manager {
	return &Manager{
		s3Client:  s3,
		stsClient: sts,
		ssmClient: ssm,
		region:    "eu-west-1",
	}
}

// ReadConfig loads the config from S3.
func (m *Manager) ReadConfig(bucket, key string) (output *Config, err error) {
	res, err := m.s3Client.GetObject(&s3.GetObjectInput{
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
func (m *Manager) WriteCredentials(creds *sts.Credentials, path, key string) error {
	values := map[string]string{
		path + "-access-key":    aws.StringValue(creds.AccessKeyId),
		path + "-secret-key":    aws.StringValue(creds.SecretAccessKey),
		path + "-session-token": aws.StringValue(creds.SessionToken),
		path + "-expiration":    creds.Expiration.Format("2006-01-02 15:04"),
	}

	for name, value := range values {
		err := m.writeSecret(name, value, key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) writeSecret(name, value, key string) error {
	input := &ssm.PutParameterInput{
		Name:      aws.String(name),
		Value:     aws.String(value),
		KeyId:     aws.String(key),
		Type:      aws.String("SecureString"),
		Overwrite: aws.Bool(true),
	}
	_, err := m.ssmClient.PutParameter(input)
	return err
}
