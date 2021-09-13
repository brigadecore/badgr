package badges

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-github/v33/github"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	s := NewService()
	require.NotNil(t, s.(*service).listCheckSuitesForRefFn)
}

func TestServiceCheckBadge(t *testing.T) {
	const testOwner = "foo"
	const testRepo = "bar"
	const testBadgeName = "build"
	testCases := []struct {
		name       string
		service    *service
		appID      int
		assertions func(CheckBadge, error)
	}{
		{
			name: "no app id; error communicating with github",
			service: &service{
				listCheckSuitesForRefFn: func(
					context.Context,
					string,
					string,
					string,
					*github.ListCheckSuiteOptions,
				) (*github.ListCheckSuiteResults, *github.Response, error) {
					return nil, nil, errors.New("something went wrong")
				},
			},
			assertions: func(_ CheckBadge, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error retrieving check suites for owner",
				)
			},
		},
		{
			name: "with app id; error communicating with github",
			service: &service{
				listCheckSuitesForRefFn: func(
					context.Context,
					string,
					string,
					string,
					*github.ListCheckSuiteOptions,
				) (*github.ListCheckSuiteResults, *github.Response, error) {
					return nil, nil, errors.New("something went wrong")
				},
			},
			appID: 42,
			assertions: func(_ CheckBadge, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error retrieving check suites for appID",
				)
			},
		},
		{
			name: "no result from github",
			service: &service{
				listCheckSuitesForRefFn: func(
					context.Context,
					string,
					string,
					string,
					*github.ListCheckSuiteOptions,
				) (*github.ListCheckSuiteResults, *github.Response, error) {
					return nil, nil, nil
				},
			},
			assertions: func(badge CheckBadge, err error) {
				require.NoError(t, err)
				require.Equal(
					t,
					CheckBadge{
						name:   testBadgeName,
						status: CheckStatusUnknown,
					},
					badge,
				)
			},
		},
		{
			name: "success",
			service: &service{
				listCheckSuitesForRefFn: func(
					context.Context,
					string,
					string,
					string,
					*github.ListCheckSuiteOptions,
				) (*github.ListCheckSuiteResults, *github.Response, error) {
					return &github.ListCheckSuiteResults{
						CheckSuites: []*github.CheckSuite{
							{
								Status:     github.String("completed"),
								Conclusion: github.String("success"),
							},
						},
					}, nil, nil
				},
			},
			assertions: func(badge CheckBadge, err error) {
				require.NoError(t, err)
				require.Equal(
					t,
					CheckBadge{
						name:   testBadgeName,
						status: CheckStatusPassed,
					},
					badge,
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(
				testCase.service.CheckBadge(
					context.Background(),
					testOwner,
					testRepo,
					&CheckBadgeOptions{
						GitHubAppID: testCase.appID,
					},
				),
			)
		})
	}
}

func TestCheckStatus(t *testing.T) {
	testCases := []struct {
		name           string
		checkSuites    []*github.CheckSuite
		expectedStatus CheckStatus
	}{
		{
			name:           "nil check results",
			checkSuites:    nil,
			expectedStatus: CheckStatusUnknown,
		},
		{
			name:           "empty check results",
			checkSuites:    []*github.CheckSuite{},
			expectedStatus: CheckStatusUnknown,
		},
		{
			name: "status completed; conclusion success",
			checkSuites: []*github.CheckSuite{
				{
					Status:     github.String("completed"),
					Conclusion: github.String("success"),
				},
			},
			expectedStatus: CheckStatusPassed,
		},
		{
			name: "status completed; conclusion failure",
			checkSuites: []*github.CheckSuite{
				{
					Status:     github.String("completed"),
					Conclusion: github.String("failure"),
				},
			},
			expectedStatus: CheckStatusFailed,
		},
		{
			name: "status completed; conclusion neutral",
			checkSuites: []*github.CheckSuite{
				{
					Status:     github.String("completed"),
					Conclusion: github.String("neutral"),
				},
			},
			expectedStatus: CheckStatusNeutral,
		},
		{
			name: "status completed; conclusion cancelled", // nolint: misspell
			// ^ This is how GitHub spells it
			checkSuites: []*github.CheckSuite{
				{
					Status:     github.String("completed"),
					Conclusion: github.String("cancelled"), // nolint: misspell
					// ^ This is how GitHub spells it
				},
			},
			expectedStatus: CheckStatusCanceled,
		},
		{
			name: "status completed; conclusion timed_out",
			checkSuites: []*github.CheckSuite{
				{
					Status:     github.String("completed"),
					Conclusion: github.String("timed_out"),
				},
			},
			expectedStatus: CheckStatusTimedOut,
		},
		{
			name: "status completed; conclusion action_required",
			checkSuites: []*github.CheckSuite{
				{
					Status:     github.String("completed"),
					Conclusion: github.String("action_required"),
				},
			},
			expectedStatus: CheckStatusActionRequired,
		},
		{
			name: "status completed; unrecognized conclusion",
			checkSuites: []*github.CheckSuite{
				{
					Status:     github.String("completed"),
					Conclusion: github.String("bogus"),
				},
			},
			expectedStatus: CheckStatusUnknown,
		},
		{
			name: "status in_progress",
			checkSuites: []*github.CheckSuite{
				{
					Status: github.String("in_progress"),
				},
			},
			expectedStatus: CheckStatusInProgress,
		},
		{
			name: "status queued",
			checkSuites: []*github.CheckSuite{
				{
					Status: github.String("queued"),
				},
			},
			expectedStatus: CheckStatusQueued,
		},
		{
			name: "unrecognized status",
			checkSuites: []*github.CheckSuite{
				{
					Status: github.String("bogus"),
				},
			},
			expectedStatus: CheckStatusUnknown,
		},
		{
			name: "two check results; same status",
			checkSuites: []*github.CheckSuite{
				{
					Status:     github.String("completed"),
					Conclusion: github.String("success"),
				},
				{
					Status:     github.String("completed"),
					Conclusion: github.String("success"),
				},
			},
			expectedStatus: CheckStatusPassed,
		},
		{
			name: "two check results; different statuses",
			checkSuites: []*github.CheckSuite{
				{
					Status:     github.String("completed"),
					Conclusion: github.String("success"),
				},
				{
					Status: github.String("in_progress"),
				},
			},
			expectedStatus: CheckStatusInProgress,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(
				t,
				testCase.expectedStatus,
				checkStatus(testCase.checkSuites),
			)
		})
	}
}
