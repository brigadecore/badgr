package badges

import "fmt"

// Color represents the color of a badge.
type Color string

const (
	// ColorGreen represents green.
	ColorGreen Color = "brightgreen"
	// ColorBlue represents blue.
	ColorBlue Color = "blue"
	// ColorYellow represents yellow.
	ColorYellow Color = "yellow"
	// ColorRed represents red.
	ColorRed Color = "red"
)

// CheckStatus represents the status of a GitHub check suite.
type CheckStatus uint8

const (
	// All the CheckStatues that follow are carefully arranged in descending order
	// of severity. In cases where multiple check suites are in play, this allows
	// Badgr to perform comparisons to find the most severe status (smallest
	// value).
	//
	// While it may seem more intuitive for more severe statuses to have HIGHER
	// values, this scheme was selected because it makes sense for the zero value
	// to equate to "unknown" and it ALSO makes sense for "unknown" to trump all
	// other statuses (i.e. to be the most severe). In this way, what we achieve
	// is that a newly minted badge with status unspecified is inherently in an
	// unknown status AND when two statuses are combined and one is unknown, the
	// combined status is unknown.

	// CheckStatusUnknown represents the case where Badgr has been unable to
	// determine a check suite's status.
	CheckStatusUnknown CheckStatus = iota
	// CheckStatusFailed represents the case where a check suite has run to
	// completion and failed.
	CheckStatusFailed
	// CheckStatusTimedOut represents the case where a check suite has time out.
	CheckStatusTimedOut
	// CheckStatusActionRequired represents the case where a check suite has run
	// to completion but some action is required from a user.
	CheckStatusActionRequired
	// CheckStatusCancelled represents the case where execution of a test suite
	// has been voluntarily terminated by a user or some other process.
	CheckStatusCanceled
	// CheckStatusNeutral represents the case where a check suite has run to
	// completion and neither failed nor succeeded.
	CheckStatusNeutral
	// CheckStatusQueued represents the case where none of the checks in the check
	// suite have been reported yet as being complete or in progress.
	CheckStatusQueued
	// CheckStatusInProgress represents the case where one or more checks in the
	// suite has progressed past the queued state, but not all checks are
	// complete.
	CheckStatusInProgress
	// CheckStatusFailed represents the case where a check suite has run to
	// completion and succeeded.
	CheckStatusPassed
)

// String returns a textual representation of a numeric CheckStatus value.
func (c CheckStatus) String() string {
	switch c {
	case CheckStatusUnknown:
		return "unknown"
	case CheckStatusFailed:
		return "failed"
	case CheckStatusTimedOut:
		return "timed out"
	case CheckStatusActionRequired:
		return "action required"
	case CheckStatusCanceled:
		return "canceled"
	case CheckStatusNeutral:
		return "neutral"
	case CheckStatusQueued:
		return "queued"
	case CheckStatusInProgress:
		return "in progress"
	case CheckStatusPassed:
		return "passed"
	default:
		return "unknown"
	}
}

// Color returns a Color constant that corresponds to the CheckStatus value.
func (c CheckStatus) Color() Color {
	switch c {
	case CheckStatusUnknown:
		return ColorYellow
	case CheckStatusFailed:
		return ColorRed
	case CheckStatusTimedOut:
		return ColorRed
	case CheckStatusActionRequired:
		return ColorYellow
	case CheckStatusCanceled:
		return ColorYellow
	case CheckStatusNeutral:
		return ColorYellow
	case CheckStatusQueued:
		return ColorBlue
	case CheckStatusInProgress:
		return ColorBlue
	case CheckStatusPassed:
		return ColorGreen
	default:
		return ColorYellow
	}
}

// Badge is an interface for any type that represents a badge.
type Badge interface {
	// Name returns the label text that should appear on the left of a badge.
	Name() string
	// Status returns the badge status that should appear on the right of a badge.
	Status() string
	// Color returns a Color constant that indicates badge color.
	Color() Color
}

type CheckBadgeOptions struct {
	// BadgeName specifies a name that should be applied to the badge. If left
	// unspecified, it will default to "build".
	BadgeName string
	// Branch indicates the branch upon whose check suites the badge should be
	// based. If left unspecified, it will default to "main".
	Branch string
	// GitHubAppID specifies that the badge should be based on the results only of
	// check suites associated with the indicated GitHub App ID. If left
	// unspecified (0), the badge will reflect the combined results of multiple
	// check suites.
	GitHubAppID int
}

// CheckBadge is an implementation of Badge that represents the results of a
// GitHub check suite, or possibly the combined results of multiple GitHub check
// suites.
type CheckBadge struct {
	name   string
	status CheckStatus
}

func (c CheckBadge) Name() string {
	return c.name
}

func (c CheckBadge) Status() string {
	return c.status.String()
}

func (c CheckBadge) Color() Color {
	return c.status.Color()
}

// ErrBadge is an implementation of a Badge that represents a Badgr failure, as
// opposed to a failure in whatever backend system (for instance, GitHub) that
// Badgr has queried.
type ErrBadge struct {
	status string
}

// NewErrBadge return an implementation of a Badge that represents a Badgr
// failure, as opposed to a failure in whatever backend system (for instance,
// GitHub) that Badgr has queried.
func NewErrBadge(status interface{}) ErrBadge {
	return ErrBadge{
		status: fmt.Sprintf("%v", status),
	}
}

func (e ErrBadge) Name() string {
	return "error"
}

func (e ErrBadge) Status() string {
	return e.status
}

func (e ErrBadge) Color() Color {
	return ColorRed
}
