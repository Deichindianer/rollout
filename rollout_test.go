package rollout_test

import (
	"errors"
	"github.com/Deichindianer/rollout"
	"strconv"
	"testing"
)

type TestService struct {
	ShouldFailRollout     bool
	ShouldFailCheckHealth bool
	ShouldFailRollback    bool
}

var ErrRollout = errors.New("rollout failed")

func (t TestService) Rollout() error {
	if t.ShouldFailRollout {
		return ErrRollout
	}
	return nil
}

var ErrCheckHealth = errors.New("check health failed")

func (t TestService) CheckHealth() error {
	if t.ShouldFailCheckHealth {
		return ErrCheckHealth
	}
	return nil
}

var ErrRollback = errors.New("rollback failed")

func (t TestService) Rollback() error {
	if t.ShouldFailRollback {
		return ErrRollback
	}
	return nil
}

func TestServiceRollout(t *testing.T) {
	testData := []struct {
		Name        string
		TestService TestService
		Errors      []error
	}{
		{
			Name:        "Happy path",
			TestService: TestService{},
			Errors:      []error{nil},
		},
		{
			Name:        "Broken rollout",
			TestService: TestService{ShouldFailRollout: true},
			Errors:      []error{ErrRollout},
		},
		{
			Name:        "Broken health check",
			TestService: TestService{ShouldFailCheckHealth: true},
			Errors:      []error{ErrCheckHealth},
		},
		{
			Name:        "Broken rollback",
			TestService: TestService{ShouldFailRollback: true},
			Errors:      []error{nil},
		},
		{
			Name: "Broken rollout and rollback",
			TestService: TestService{
				ShouldFailRollout:  true,
				ShouldFailRollback: true,
			},
			Errors: []error{ErrRollout, ErrRollback},
		},
		{
			Name: "Broken health check and rollback",
			TestService: TestService{
				ShouldFailCheckHealth: true,
				ShouldFailRollback:    true,
			},
			Errors: []error{ErrCheckHealth, ErrRollback},
		},
		{
			Name: "Broken rollout, health check and rollback",
			TestService: TestService{
				ShouldFailRollout:     true,
				ShouldFailCheckHealth: true,
				ShouldFailRollback:    true,
			},
			Errors: []error{ErrRollout, ErrRollback},
		},
	}

	for _, td := range testData {
		t.Run(td.Name, func(t *testing.T) {
			err := rollout.ServiceRollout(td.TestService)

			for _, e := range td.Errors {
				if err == nil && e != nil {
					t.Errorf("expected %s but got %s error", e, err)
				}

				if !errors.Is(err, e) {
					t.Errorf("expected `%s` error but got `%s`", e, err)
				}
			}
		})
	}
}

func TestServiceErr_Error(t *testing.T) {
	testData := []struct {
		Name         string
		ServiceError rollout.ServiceErr
		ExpectedMsg  string
	}{
		{
			Name:         "Rollout Error",
			ServiceError: rollout.ServiceErr{RolloutErr: errors.New("rollout failed")},
			ExpectedMsg:  "failed rollout: rollout failed",
		},
		{
			Name: "Check health Rollback Msg",
			ServiceError: rollout.ServiceErr{
				RollbackSuccessful: true,
				RolloutErr:         errors.New("rollout failed"),
			},
			ExpectedMsg: "rollback successful: failed rollout: rollout failed",
		},
		{
			Name: "Check Health Error",
			ServiceError: rollout.ServiceErr{
				CheckHealthErr: errors.New("check health failed"),
			},
			ExpectedMsg: "failed health check: check health failed",
		},
		{
			Name: "Check health Rollback Msg",
			ServiceError: rollout.ServiceErr{
				RollbackSuccessful: true,
				CheckHealthErr:     errors.New("check health failed"),
			},
			ExpectedMsg: "rollback successful: failed health check: check health failed",
		},
		{
			Name: "Rollout Rollback Error",
			ServiceError: rollout.ServiceErr{
				RolloutErr:  errors.New("rollout failed"),
				RollbackErr: errors.New("rollback failed"),
			},
			ExpectedMsg: "failed rollout: rollout failed: failed rollback: rollback failed",
		},
		{
			Name: "Check health Rollback Error",
			ServiceError: rollout.ServiceErr{
				CheckHealthErr: errors.New("check health failed"),
				RollbackErr:    errors.New("rollback failed"),
			},
			ExpectedMsg: "failed health check: check health failed: failed rollback: rollback failed",
		},
		{
			Name:         "Empty ServiceErr",
			ServiceError: rollout.ServiceErr{},
			ExpectedMsg:  "",
		},
	}

	for _, td := range testData {
		t.Run(td.Name, func(t *testing.T) {
			msg := td.ServiceError.Error()
			if msg != td.ExpectedMsg {
				t.Errorf("expected `%s` got `%s`", td.ExpectedMsg, msg)
			}
		})
	}
}

func TestServiceErr_Is(t *testing.T) {
	testData := []struct {
		Name          string
		ServiceError  rollout.ServiceErr
		TargetErr     error
		ExpectedMatch bool
	}{
		{
			Name:         "RolloutErrMatch",
			ServiceError: rollout.ServiceErr{RolloutErr: errors.New("test")},
			TargetErr:    errors.New("test"),
		},
		{
			Name:         "CheckHealthErrMatch",
			ServiceError: rollout.ServiceErr{CheckHealthErr: errors.New("test")},
			TargetErr:    errors.New("test"),
		},
		{
			Name:         "RollbackErrMatch",
			ServiceError: rollout.ServiceErr{RollbackErr: errors.New("test")},
			TargetErr:    errors.New("test"),
		},
		{
			Name:         "NoMatch",
			ServiceError: rollout.ServiceErr{RollbackErr: errors.New("test")},
			TargetErr:    errors.New("not test"),
		},
		{
			Name:         "ServiceErrNil",
			ServiceError: rollout.ServiceErr{},
			TargetErr:    errors.New("not test"),
		},
		{
			Name:         "TargetErrNil",
			ServiceError: rollout.ServiceErr{},
			TargetErr:    nil,
		},
	}

	for _, td := range testData {
		t.Run(td.Name, func(t *testing.T) {
			match := errors.Is(td.ServiceError, td.TargetErr)

			if match != td.ExpectedMatch {
				t.Errorf("expected %s but got %s",
					strconv.FormatBool(td.ExpectedMatch),
					strconv.FormatBool(match),
				)
			}
		})
	}
}
