## concourse-sts-lambda

[![Build Status](https://travis-ci.org/telia-oss/concourse-sts-lambda.svg?branch=master)](https://travis-ci.org/telia-oss/concourse-sts-lambda)

Lambda function to rotate AWS credentials used by Concourse teams. See 
the terraform subdirectory for an example that should work (with minimal effort).

### Why?

Our CI/CD (in our case Concourse) needs AWS credentials to deploy Terraform
templates. Since we are sharing workers between teams, the instance profile
itself has no privileges. And so, we need to pass in credentials to the tasks 
which require them.

Instead of having individual teams being responsible for their CI credentials,
we can use this Lambda function to write temporary credentials to a specific teams
Concourse secrets, for one or more accounts.

### How?

In short:

1. This Lambda function is deployed to the same account as our Concourse.
2. Individual accounts add a CI role with the Lambda functions execution role
as a trusted entity.
3. A team adds a CloudWatch event rule with the configuration for which
accounts they need access to.
4. Lambda assumes the roles specified in the configuration and rotates 
the temporary AWS credentials for said team on a 50min schedule.
5. ???
6. Profit.

### Usage

Be in the root directory:

```bash
make release
```

You should now have a zipped Lambda function. Next, edit [terraform/example.tf](./terraform/example.tf)
to your liking. When done, be in the terraform directory:

```bash
terraform init
terraform apply
```

NOTE: The `aws/secretsmanager` KMS Key Alias has to be created/exist before the lambda is deployed.

### Team configuration

Example configuration for a Team (which is then passed as input in the CloudWatch event rule):

```json
{
  "name": "example-team",
  "accounts": [{
    "name": "divx-lab",
    "roleArn": "arn:aws:iam::123456789999:role/machine-user-example"
  }]
}
```

You can also optionally specify the duration the sts assume role call should use:

```json
{
  "name": "example-team",
  "accounts": [{
    "name": "divx-lab",
    "roleArn": "arn:aws:iam::123456789999:role/machine-user-example",
    "duration": 7200
  }]
}
```

Note: [Role chaining](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_terms-and-concepts.html) has a limit on duration to be maximum 1 hour.

When the function is triggered with this input it will assume the
`roleArn`, and write the credentials to (by default):

- `/concourse/example-team/divx-lab-access-key`
- `/concourse/example-team/divx-lab-secret-key`
- `/concourse/example-team/divx-lab-session-token`

Note that you can have multiple accounts, in which case the account
name must be unique to avoid overwriting the secrets in Secrets manager.
