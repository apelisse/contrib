/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package matchers

// Matcher is an interface to match an event
import (
	"strings"
	"time"

	"github.com/google/go-github/github"
)

// Matcher matches against a comment or an event
type Matcher interface {
	MatchEvent(event *github.IssueEvent) bool
	MatchComment(comment *github.IssueComment) bool
	MatchReviewComment(comment *github.PullRequestComment) bool
}

// CreatedAfter matches comments created after the time
type CreatedAfter time.Time

var _ Matcher = CreatedAfter{}

// MatchComment returns true if the comment is created after the time
func (c CreatedAfter) MatchComment(comment *github.IssueComment) bool {
	if comment == nil || comment.CreatedAt == nil {
		return false
	}
	return comment.CreatedAt.After(time.Time(c))
}

// MatchEvent returns true if the event is created after the time
func (c CreatedAfter) MatchEvent(event *github.IssueEvent) bool {
	if event == nil || event.CreatedAt == nil {
		return false
	}
	return event.CreatedAt.After(time.Time(c))
}

// MatchReviewComment returns true if the review comment is created after the time
func (c CreatedAfter) MatchReviewComment(review *github.PullRequestComment) bool {
	if review == nil || review.CreatedAt == nil {
		return false
	}
	return review.CreatedAt.After(time.Time(c))
}

// CreatedBefore matches Items created before the time
type CreatedBefore time.Time

var _ Matcher = CreatedBefore{}

// MatchComment returns true if the comment is created before the time
func (c CreatedBefore) MatchComment(comment *github.IssueComment) bool {
	if comment == nil || comment.CreatedAt == nil {
		return false
	}
	return comment.CreatedAt.Before(time.Time(c))
}

// MatchEvent returns true if the event is created before the time
func (c CreatedBefore) MatchEvent(event *github.IssueEvent) bool {
	if event == nil || event.CreatedAt == nil {
		return false
	}
	return event.CreatedAt.Before(time.Time(c))
}

// MatchReviewComment returns true if the review comment is created before the time
func (c CreatedBefore) MatchReviewComment(review *github.PullRequestComment) bool {
	if review == nil || review.CreatedAt == nil {
		return false
	}
	return review.CreatedAt.Before(time.Time(c))
}

type ValidAuthor struct{}

var _ Matcher = ValidAuthor{}

func (v ValidAuthor) MatchEvent(event *github.IssueEvent) bool {
	return event != nil && event.Actor != nil && event.Actor.Login != nil
}

func (v ValidAuthor) MatchComment(comment *github.IssueComment) bool {
	return comment != nil && comment.User != nil && comment.User.Login != nil
}

func (v ValidAuthor) MatchReviewComment(review *github.PullRequestComment) bool {
	return review != nil && review.User != nil && review.User.Login != nil
}

type AuthorLogin string

var _ Matcher = AuthorLogin("")

func (a AuthorLogin) MatchEvent(event *github.IssueEvent) bool {
	if !(ValidAuthor{}).MatchEvent(event) {
		return false
	}

	return strings.ToLower(*event.Actor.Login) == strings.ToLower(string(a))
}

func (a AuthorLogin) MatchComment(comment *github.IssueComment) bool {
	if !(ValidAuthor{}).MatchComment(comment) {
		return false
	}

	return strings.ToLower(*comment.User.Login) == strings.ToLower(string(a))
}

func (a AuthorLogin) MatchReviewComment(review *github.PullRequestComment) bool {
	if !(ValidAuthor{}).MatchReviewComment(review) {
		return false
	}

	return strings.ToLower(*review.User.Login) == strings.ToLower(string(a))
}

func AuthorLogins(authors ...string) Matcher {
	or := OrMatcher{}

	for _, author := range authors {
		or = append(or, AuthorLogin(author))
	}

	return or
}

func AuthorUsers(users ...*github.User) Matcher {
	authors := []string{}

	for _, user := range users {
		if user == nil || user.Login == nil {
			continue
		}
		authors = append(authors, *user.Login)
	}

	return AuthorLogins(authors...)
}

// AddLabel searches for "labeled" event.
type AddLabel struct{}

// Match if the event is of type "labeled"
func (a AddLabel) MatchEvent(event *github.IssueEvent) bool {
	if event == nil || event.Event == nil {
		return false
	}
	return *event.Event == "labeled"
}

func (a AddLabel) MatchComment(comment *github.IssueComment) bool {
	return false
}

func (a AddLabel) MatchReviewComment(review *github.PullRequestComment) bool {
	return false
}

// LabelPrefix searches for event whose label starts with the string
type LabelPrefix string

// Match if the label starts with the string
func (l LabelPrefix) MatchEvent(event *github.IssueEvent) bool {
	if event == nil || event.Label == nil || event.Label.Name == nil {
		return false
	}
	return strings.HasPrefix(*event.Label.Name, string(l))
}

func (l LabelPrefix) MatchComment(comment *github.IssueComment) bool {
	return false
}

func (l LabelPrefix) MatchReviewComment(review *github.PullRequestComment) bool {
	return false
}

type EventType struct{}

var _ Matcher = EventType{}

func (c EventType) MatchEvent(event *github.IssueEvent) bool {
	return true
}

func (c EventType) MatchComment(comment *github.IssueComment) bool {
	return false
}

func (c EventType) MatchReviewComment(review *github.PullRequestComment) bool {
	return false
}

type CommentType struct{}

var _ Matcher = CommentType{}

func (c CommentType) MatchEvent(event *github.IssueEvent) bool {
	return false
}

func (c CommentType) MatchComment(comment *github.IssueComment) bool {
	return true
}

func (c CommentType) MatchReviewComment(review *github.PullRequestComment) bool {
	return false
}

type ReviewCommentType struct{}

var _ Matcher = ReviewCommentType{}

func (c ReviewCommentType) MatchEvent(event *github.IssueEvent) bool {
	return false
}

func (c ReviewCommentType) MatchComment(comment *github.IssueComment) bool {
	return false
}

func (c ReviewCommentType) MatchReviewComment(review *github.PullRequestComment) bool {
	return true
}
