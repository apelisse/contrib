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

package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

// Client can be used to run commands again Github API
type Client struct {
	Token     string
	TokenFile string
	Org       string
	Project   string

	githubClient *github.Client
}

const (
	tokenLimit = 50 // We try to stop that far from the API limit
)

// AddFlags parses options for github client
func (client *Client) AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&client.Token, "token", "",
		"The OAuth Token to use for requests.")
	cmd.PersistentFlags().StringVar(&client.TokenFile, "token-file", "",
		"The file containing the OAuth Token to use for requests.")
	cmd.PersistentFlags().StringVar(&client.Org, "organization", "kubernetes",
		"The github organization to scan")
	cmd.PersistentFlags().StringVar(&client.Project, "project", "kubernetes",
		"The github project to scan")
}

// Create the github client that we use to communicate with github
func (client *Client) getGithubClient() (*github.Client, error) {
	if client.githubClient != nil {
		return client.githubClient, nil
	}
	token := client.Token
	if len(token) == 0 && len(client.TokenFile) != 0 {
		data, err := ioutil.ReadFile(client.TokenFile)
		if err != nil {
			return nil, err
		}
		token = strings.TrimSpace(string(data))
	}

	if len(token) > 0 {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		client.githubClient = github.NewClient(tc)
	} else {
		client.githubClient = github.NewClient(nil)
	}
	return client.githubClient, nil
}

// Make sure we have not reached the limit or wait
func (client *Client) limitsCheckAndWait() {
	var sleep time.Duration
	githubClient, err := client.getGithubClient()
	if err != nil {
		glog.Errorf("Failed to get RateLimits: %v", err)
		sleep = time.Minute
	} else {
		limits, _, err := githubClient.RateLimits()
		if err != nil {
			glog.Errorf("Failed to get RateLimits: %v", err)
			sleep = time.Minute
		}
		if limits != nil && limits.Core != nil && limits.Core.Remaining < tokenLimit {
			sleep = limits.Core.Reset.Sub(time.Now())
			glog.Infof("RateLimits: reached. Sleeping for %v", sleep)
		}
	}

	time.Sleep(sleep)
}

// ClientInterface describes what a client should be able to do
type ClientInterface interface {
	FetchIssues(time.Time) ([]github.Issue, error)
	FetchIssueEvents(*int) ([]github.IssueEvent, error)
}

// FetchIssues from Github, until 'latest' time
func (client *Client) FetchIssues(latest time.Time) ([]github.Issue, error) {
	var allIssues []github.Issue
	opt := &github.IssueListByRepoOptions{Since: latest, Sort: "updated", State: "all", Direction: "asc"}

	githubClient, err := client.getGithubClient()
	if err != nil {
		return nil, err
	}

	for {
		client.limitsCheckAndWait()

		issues, resp, err := githubClient.Issues.ListByRepo(client.Org, client.Project, opt)
		if err != nil {
			return nil, err
		}

		for _, issue := range issues {
			fmt.Println("Issue", *issue.Number, "last updated", *issue.UpdatedAt)
		}

		allIssues = append(allIssues, issues...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	return allIssues, nil
}

// Look for a specific Id in a list of events
func wasIdFound(events []github.IssueEvent, id int) bool {
	for _, event := range events {
		if *event.ID == id {
			return true
		}
	}
	return false
}

// FetchIssueEvents from github and return the full list, until it matches 'latest'
// The entire last page will be included so you can have redundancy.
func (client *Client) FetchIssueEvents(latest *int) ([]github.IssueEvent, error) {
	var allEvents []github.IssueEvent
	opt := &github.ListOptions{PerPage: 100}

	githubClient, err := client.getGithubClient()
	if err != nil {
		return nil, err
	}

	for {
		client.limitsCheckAndWait()

		fmt.Println("Downloading events page: ", opt.Page)
		events, resp, err := githubClient.Issues.ListRepositoryEvents(client.Org, client.Project, opt)
		if err != nil {
			glog.Errorf("Request failed. Wait before trying again.")
			time.Sleep(time.Second)
			continue
		}

		allEvents = append(allEvents, events...)
		if resp.NextPage == 0 || (latest != nil && wasIdFound(events, *latest)) {
			break
		}
		break
		opt.Page = resp.NextPage
	}

	return allEvents, nil
}