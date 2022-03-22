// Package rollout defines a common way of safely rolling out
// most generic applications, and a small set of pre-defined
// flows for common patterns.
package rollout

import (
	"errors"
)

type Service interface {
	Rollout() error
	CheckHealth() error
	Rollback() error
}

type ServiceErr struct {
	RollbackSuccessful bool
	RolloutErr         error
	CheckHealthErr     error
	RollbackErr        error
}

func (se ServiceErr) Error() string {
	errorMsg := ""

	if se.RolloutErr != nil {
		errorMsg = "failed rollout: " + se.RolloutErr.Error()
	}

	if se.CheckHealthErr != nil {
		errorMsg += "failed health check: " + se.CheckHealthErr.Error()
	}

	if se.RollbackErr != nil {
		errorMsg += ": failed rollback: " + se.RollbackErr.Error()
	}

	if se.RollbackSuccessful {
		return "rollback successful: " + errorMsg
	}
	return errorMsg
}

func (se ServiceErr) Is(target error) bool {
	if errors.Is(target, se.RolloutErr) {
		return true
	}

	if errors.Is(target, se.CheckHealthErr) {
		return true
	}

	if errors.Is(target, se.RollbackErr) {
		return true
	}

	return false
}

func ServiceRollout(s Service) error {
	err := s.Rollout()
	if err != nil {
		rollbackErr := s.Rollback()
		if rollbackErr != nil {
			return ServiceErr{
				RolloutErr:  err,
				RollbackErr: rollbackErr,
			}
		}

		return ServiceErr{
			RollbackSuccessful: true,
			RolloutErr:         err,
		}
	}

	err = s.CheckHealth()
	if err != nil {
		rollbackErr := s.Rollback()
		if rollbackErr != nil {
			return ServiceErr{
				CheckHealthErr: err,
				RollbackErr:    rollbackErr,
			}
		}

		return ServiceErr{
			RollbackSuccessful: true,
			CheckHealthErr:     err,
		}
	}
	return nil
}
