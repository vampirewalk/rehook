package main

import (
	"log"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const paginage = false

type IssueAPI struct {
	Client *github.Client
}

func NewIssueAPI() *IssueAPI {
	token, err := queryToken("Github_Token", "vampirewalk")
	if err != nil {
		client := github.NewClient(nil)
		return &IssueAPI{Client: client}
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	return &IssueAPI{Client: client}
}

func (i IssueAPI) ListIssues(user, repo string) ([]github.Issue, error) {
	opts := &github.IssueListByRepoOptions{
		Sort: "created",
	}

	allIssues := make([]github.Issue, 0)

	for {
		log.Println("querying github for issues: page", opts.ListOptions.Page)
		iss, resp, err := i.Client.Issues.ListByRepo(user, repo, opts)
		if err != nil {
			return nil, err
		}

		allIssues = append(allIssues, iss...)

		/*for _, i := range iss {
			fmt.Println(i)
		}*/
		if !paginage || resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}
	return allIssues, nil
}

func (i IssueAPI) ListIssueEvents(user, repo string, issueNum int) ([]github.IssueEvent, error) {
	opts := &github.ListOptions{Page: 0}

	allEvents := make([]github.IssueEvent, 0)

	for {
		events, resp, err := i.Client.Issues.ListIssueEvents(user, repo, issueNum, opts)
		if err != nil {
			return nil, err
		}

		allEvents = append(allEvents, events...)

		/*for _, e := range events {
			fmt.Println(e)
		}*/
		if !paginage || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage

	}
	return allEvents, nil
}
