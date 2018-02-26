package main_test

import (
	"errors"
	pkg "github.com/TeliaSoneraNorge/concourse-sts-lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type MockS3 struct {
	s3iface.S3API
	Error bool
}

type MockSTS struct {
	stsiface.STSAPI
	Error bool
}

func (mock *MockSTS) AssumeRole(input *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
	if mock.Error == true {
		return nil, errors.New("expected")
	}
	if input.DurationSeconds == nil {
		return nil, errors.New("missing DurationSeconds")
	}
	if input.RoleArn == nil {
		return nil, errors.New("missing RoleArn")
	}
	if input.RoleSessionName == nil {
		return nil, errors.New("missing RoleSessionName")
	}

	return &sts.AssumeRoleOutput{
		Credentials: &sts.Credentials{
			AccessKeyId:     aws.String("accesskey"),
			SecretAccessKey: aws.String("secretkey"),
			SessionToken:    aws.String("sessiontoken"),
			Expiration:      aws.Time(time.Date(2018, time.January, 27, 13, 32, 52, 0, time.UTC)),
		},
	}, nil
}

type MockSSM struct {
	ssmiface.SSMAPI
	Error bool
}

func (mock *MockSSM) PutParameter(input *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	if mock.Error == true {
		return nil, errors.New("expected")
	}
	if input.KeyId == nil {
		return nil, errors.New("missing KeyId")
	}
	if input.Overwrite == nil {
		return nil, errors.New("missing Overwrite")
	}
	if input.Name == nil {
		return nil, errors.New("missing Name")
	}
	if input.Value == nil {
		return nil, errors.New("missing Value")
	}
	if input.Type == nil || aws.StringValue(input.Type) != "SecureString" {
		return nil, errors.New("missing Type or type is not 'SecureString'")
	}

	return &ssm.PutParameterOutput{
		Version: aws.Int64(1),
	}, nil
}

var (
	account = &pkg.Account{
		Name:    "account",
		RoleArn: "role",
	}
	team = &pkg.Team{
		Name:     "team",
		KeyID:    "key",
		Accounts: []*pkg.Account{account},
	}
)

func TestAssume(t *testing.T) {
	ssmMock := &MockSSM{
		Error: false,
	}
	stsMock := &MockSTS{
		Error: false,
	}

	m := pkg.NewTestManager(stsMock, ssmMock)

	t.Run("Assume works", func(t *testing.T) {
		expected := &sts.Credentials{
			AccessKeyId:     aws.String("accesskey"),
			SecretAccessKey: aws.String("secretkey"),
			SessionToken:    aws.String("sessiontoken"),
			Expiration:      aws.Time(time.Date(2018, time.January, 27, 13, 32, 52, 0, time.UTC)),
		}

		actual, err := m.AssumeRole(account.RoleArn, team.Name)
		assert.Nil(t, err)
		assert.NotNil(t, actual)
		assert.Equal(t, expected, actual)
	})

	t.Run("Errors are propagated", func(t *testing.T) {
		stsMock.Error = true
		defer func() {
			stsMock.Error = false
		}()

		actual, err := m.AssumeRole(account.RoleArn, team.Name)
		assert.NotNil(t, err)
		assert.EqualError(t, err, "expected")
		assert.Nil(t, actual)
	})
}

func TestWriteCredentials(t *testing.T) {
	ssmMock := &MockSSM{
		Error: false,
	}
	stsMock := &MockSTS{
		Error: false,
	}

	m := pkg.NewTestManager(stsMock, ssmMock)

	input := &sts.Credentials{
		AccessKeyId:     aws.String("accesskey"),
		SecretAccessKey: aws.String("secretkey"),
		SessionToken:    aws.String("sessiontoken"),
		Expiration:      aws.Time(time.Date(2018, time.January, 27, 13, 32, 52, 0, time.UTC)),
	}

	t.Run("WriteCredentials work", func(t *testing.T) {
		err := m.WriteCredentials(input, "path", "key")
		assert.Nil(t, err)
	})

	t.Run("Errors are propagated", func(t *testing.T) {
		ssmMock.Error = true
		defer func() {
			ssmMock.Error = false
		}()

		err := m.WriteCredentials(input, "path", "key")
		assert.NotNil(t, err)
		assert.EqualError(t, err, "expected")
	})
}
