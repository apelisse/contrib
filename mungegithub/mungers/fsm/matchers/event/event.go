package event

import (
	"time"

	"github.com/google/go-github/github"
)

type Matcher interface {
	Match(event *github.IssueEvent) bool
}

type And []Matcher

func (a And) Match(event *github.IssueEvent) bool {
	for _, matcher := range []Matcher(a) {
		if !matcher.Match(event) {
			return false
		}
	}
	return true
}

type Or []Matcher

func (o Or) Match(event *github.IssueEvent) bool {
	for _, matcher := range []Matcher(o) {
		if matcher.Match(event) {
			return true
		}
	}
	return false
}

type Not struct {
	Matcher Matcher
}

func (n Not) Match(event *github.IssueEvent) bool {
	return !n.Matcher.Match(event)
}

type Actor string

func (a Actor) Match(event *github.IssueEvent) bool {
	if event.Actor == nil || event.Actor.Login == nil {
		return false
	}
	return *event.Actor.Login == string(a)
}

type AddLabel string

func (a AddLabel) Match(event *github.IssueEvent) bool {
	if event.Label == nil || event.Label.Name == nil || event.Event == nil {
		return false
	}
	return *event.Event == "labeled" || *event.Label.Name == string(a)
}

type CreatedAfter time.Time

func (c CreatedAfter) Match(event *github.IssueEvent) bool {
	return event.CreatedAt.After(time.Time(c))
}

type CreatedBefore time.Time

func (c CreatedBefore) Match(event *github.IssueEvent) bool {
	return event.CreatedAt.Before(time.Time(c))
}
