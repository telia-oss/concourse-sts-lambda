package main

// Config is the configuration for the STS Lambda
type Config []*Team

// Team represents the configuration for a single CI/CD team.
type Team struct {
	Name     string     `json:"name"`
	KeyID    string     `json:"keyId"`
	Accounts []*Account `json:"accounts"`
}

// Account represents the configuration for an assumable role.
type Account struct {
	Name    string `json:"name"`
	RoleArn string `json:"roleArn"`
}
