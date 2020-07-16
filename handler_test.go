package handler_test

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/golang/mock/gomock"
	logrus "github.com/sirupsen/logrus/hooks/test"
	"github.com/telia-oss/concourse-sts-lambda"
	"github.com/telia-oss/concourse-sts-lambda/mocks"
)

func TestHandler(t *testing.T) {
	config := handler.Configuration{
		Bucket: "bucket",
		Key:    "key",
	}

	team := strings.TrimSpace(`
{
	"name": "test-team",
	"accounts": [{
		"name": "test-account",
		"roleArn": "test-account-arn"
	}],
	"secretTags": {
		"Team": "test-team"
	}
}
	`)

	creds := &sts.AssumeRoleOutput{
		Credentials: &sts.Credentials{
			AccessKeyId:     aws.String("access-key"),
			SecretAccessKey: aws.String("secret-key"),
			SessionToken:    aws.String("token"),
			Expiration:      aws.Time(time.Now()),
		},
	}

	tests := []struct {
		description       string
		path              string
		config            handler.Configuration
		team              string
		stsOutput         *sts.AssumeRoleOutput
		stsError          error
		putSecretError    error
		createSecretError error
	}{

		{
			description:       "assumes a role and writes the secrets",
			path:              "/concourse/{{.Team}}/{{.Account}}",
			config:            config,
			team:              team,
			stsOutput:         creds,
			stsError:          nil,
			createSecretError: nil,
			putSecretError:    nil,
		},

		{
			description:       "continues if it fails to assume role",
			path:              "/concourse/{{.Team}}/{{.Account}}",
			config:            config,
			team:              team,
			stsOutput:         nil,
			stsError:          errors.New("test-error"),
			createSecretError: nil,
			putSecretError:    nil,
		},

		{
			description:       "continues if it fails create secret",
			path:              "/concourse/{{.Team}}/{{.Account}}",
			config:            config,
			team:              team,
			stsOutput:         creds,
			stsError:          nil,
			createSecretError: errors.New("test-error"),
			putSecretError:    nil,
		},

		{
			description:       "continues if it fails write secret",
			path:              "/concourse/{{.Team}}/{{.Account}}",
			config:            config,
			team:              team,
			stsOutput:         creds,
			stsError:          nil,
			createSecretError: nil,
			putSecretError:    errors.New("test-error"),
		},

		{
			description:       "does not error if the secret already exists",
			path:              "/concourse/{{.Team}}/{{.Account}}",
			config:            config,
			team:              team,
			stsOutput:         creds,
			stsError:          nil,
			createSecretError: awserr.New(secretsmanager.ErrCodeResourceExistsException, "", errors.New("test-error")),
			putSecretError:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s3Object := &s3.GetObjectOutput{
				Body: ioutil.NopCloser(strings.NewReader(tc.team)),
			}

			s3 := mocks.NewMockS3Client(ctrl)
			s3.EXPECT().GetObject(gomock.Any()).Times(1).Return(s3Object, nil)

			sts := mocks.NewMockSTSClient(ctrl)
			sts.EXPECT().AssumeRole(gomock.Any()).Times(1).Return(tc.stsOutput, tc.stsError)

			secrets := mocks.NewMockSecretsClient(ctrl)
			if tc.stsError == nil {
				secrets.EXPECT().CreateSecret(gomock.Any()).MinTimes(1).Return(nil, tc.createSecretError)
			}
			if tc.stsError == nil {
				if tc.createSecretError != nil {
					if e, ok := tc.createSecretError.(awserr.Error); ok {
						if e.Code() == secretsmanager.ErrCodeResourceExistsException {
							secrets.EXPECT().UpdateSecret(gomock.Any()).MinTimes(1).Return(nil, tc.putSecretError)
						}
					}
				} else {
					secrets.EXPECT().UpdateSecret(gomock.Any()).MinTimes(1).Return(nil, tc.putSecretError)
				}
			}

			logger, _ := logrus.NewNullLogger()
			handle := handler.New(handler.NewTestManager(secrets, sts, s3), tc.path, logger)

			if err := handle(tc.config); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}
