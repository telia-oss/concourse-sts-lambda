package main

import (
	"strings"
	"text/template"
)

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

// NewPath a new secret path...
func NewPath(team, account, template string) *Path {
	return &Path{
		Team:     team,
		Account:  account,
		Template: template,
	}
}

// Path represents the path used to write secrets into SSM.
type Path struct {
	Team     string
	Account  string
	Template string
}

func (p *Path) String() (string, error) {
	t, err := template.New("path").Option("missingkey=error").Parse(p.Template)
	if err != nil {
		return "", err
	}
	var s strings.Builder
	if err = t.Execute(&s, p); err != nil {
		return "", err
	}
	return s.String(), nil
}
