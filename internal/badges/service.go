package badges

import (
	"context"
	"math"

	"github.com/google/go-github/v33/github"
	"github.com/pkg/errors"
)

// Service is an interface for components that can handle requests for a badge.
type Service interface {
	// CheckBadge serves a badge based on check suite status.
	CheckBadge(
		ctx context.Context,
		owner string,
		repo string,
		opts *CheckBadgeOptions,
	) (CheckBadge, error)
}

type service struct {
	// The following functions are usually provided by a GitHub client, but are
	// overridable for testing purposes
	listCheckSuitesForRefFn func(
		ctx context.Context,
		owner string,
		repo string,
		ref string,
		opt *github.ListCheckSuiteOptions,
	) (*github.ListCheckSuiteResults, *github.Response, error)
}

// NewService returns an implementation of the Service interface for handling
// requests for a badge.
func NewService() Service {
	return &service{
		listCheckSuitesForRefFn: github.NewClient(nil).Checks.ListCheckSuitesForRef,
	}
}

func (s *service) CheckBadge(
	ctx context.Context,
	owner string,
	repo string,
	opts *CheckBadgeOptions,
) (CheckBadge, error) {
	if opts == nil {
		opts = &CheckBadgeOptions{}
	}
	if opts.BadgeName == "" {
		opts.BadgeName = "build"
	}
	if opts.Branch == "" {
		opts.Branch = "main"
	}

	badge := CheckBadge{
		name:   opts.BadgeName,
		status: CheckStatusUnknown,
	}

	checkSuites := []*github.CheckSuite{}
	ghOpts := &github.ListCheckSuiteOptions{
		ListOptions: github.ListOptions{
			Page: 1,
		},
	}
	if opts.GitHubAppID != 0 {
		ghOpts.AppID = &opts.GitHubAppID
	}
	for {
		results, response, err :=
			s.listCheckSuitesForRefFn(ctx, owner, repo, opts.Branch, ghOpts)
		if err != nil {
			if opts.GitHubAppID == 0 {
				return badge, errors.Wrapf(
					err,
					"error retrieving check suites for owner %q, repo %q, branch %q "+
						"from GitHub",
					owner,
					repo,
					opts.Branch,
				)
			}
			return badge, errors.Wrapf(
				err,
				"error retrieving check suites for appID %d, owner %q, repo %q, "+
					"branch %q from GitHub",
				opts.GitHubAppID,
				owner,
				repo,
				opts.Branch,
			)
		}

		if results == nil {
			break
		}
		checkSuites = append(checkSuites, results.CheckSuites...)
		if response == nil || response.NextPage == 0 {
			break
		}
		ghOpts.ListOptions.Page = response.NextPage
	}

	badge.status = checkStatus(checkSuites)
	return badge, nil
}

func checkStatus(checkSuites []*github.CheckSuite) CheckStatus {
	if len(checkSuites) == 0 {
		return CheckStatusUnknown
	}
	// To consolidate the status of many checks into a single badge, we start with
	// a passed badge, then we iterate over all checks, progressively degrading
	// the badge status if/as worse outcomes are found.
	status := CheckStatusPassed
	for _, checkSuite := range checkSuites {
		// Default to unknown if we cannot figure out the status
		newStatus := CheckStatusUnknown
		switch checkSuite.GetStatus() {
		case "completed":
			switch checkSuite.GetConclusion() {
			case "success":
				newStatus = CheckStatusPassed
			case "failure":
				newStatus = CheckStatusFailed
			case "neutral":
				newStatus = CheckStatusNeutral
			case "cancelled": // nolint: misspell
				// ^ This is how GitHub spells it
				newStatus = CheckStatusCanceled
			case "timed_out":
				newStatus = CheckStatusTimedOut
			case "action_required":
				newStatus = CheckStatusActionRequired
			}
		case "in_progress":
			newStatus = CheckStatusInProgress
		case "queued":
			newStatus = CheckStatusQueued
		}
		// The badge's new status is the higher severity of the two
		// Lower value == higher severity-- that allows unknown (0) to be treated as
		// most severe.
		status = CheckStatus(math.Min(float64(status), float64(newStatus)))
	}
	return status
}
