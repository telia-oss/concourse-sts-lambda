package handler

import (
	"github.com/sirupsen/logrus"
)

// New lambda handler with the provided settings.
func New(manager *Manager, secretTemplate string, logger *logrus.Logger) func(Team) error {
	return func(team Team) error {
		log := logger.WithFields(logrus.Fields{"team": team.Name})

		// Loop through teams and assume roles/write credentials for
		// all accounts controlled by the team.
		for _, account := range team.Accounts {
			path, err := NewSecretPath(team.Name, account.Name, secretTemplate).String()
			if err != nil {
				log.WithFields(logrus.Fields{"role": account.Name}).Warnf("failed to parse secret path: %s", err)
				continue
			}
			creds, err := manager.AssumeRole(account.RoleArn, team.Name)
			if err != nil {
				log.WithFields(logrus.Fields{"role": account.Name}).Warnf("failed to assume role: %s", err)
				continue
			}
			if err := manager.WriteCredentials(creds, path); err != nil {
				log.WithFields(logrus.Fields{"role": account.Name}).Warnf("failed to write credentials: %s", err)
				continue
			}
		}
		return nil
	}
}
