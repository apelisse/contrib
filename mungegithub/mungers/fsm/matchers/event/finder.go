package event

import "github.com/google/go-github/github"

func FindEvent(events []*github.IssueEvent, matcher Matcher) []*github.IssueEvent {
	matchingEvents := []*github.IssueEvent{}

	for _, event := range events {
		if matcher.Match(event) {
			matchingEvents = append(matchingEvents, event)
		}
	}

	return matchingEvents
}
