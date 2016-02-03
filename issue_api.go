package main

import (
	"fmt"
	"log"

	"github.com/google/go-github/github"
)

const paginage = false

type IssueAPI struct {
}

func (i IssueAPI) ListIssues(user, repo string) ([]github.Issue, error) {
	client := github.NewClient(nil)

	opts := &github.IssueListByRepoOptions{
		Sort: "created",
	}

	allIssues := make([]github.Issue, 0)

	for {
		log.Println("querying github for issues: page", opts.ListOptions.Page)
		iss, resp, err := client.Issues.ListByRepo(user, repo, opts)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		allIssues = append(allIssues, iss...)

		for _, i := range iss {
			fmt.Println(i)
		}
		if !paginage || resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}
	return allIssues, nil
}

func (i IssueAPI) ListIssueEvents(user, repo string, issueNum int) ([]github.IssueEvent, error) {
	client := github.NewClient(nil)

	opts := &github.ListOptions{Page: 0}

	allEvents := make([]github.IssueEvent, 0)

	for {
		events, resp, err := client.Issues.ListIssueEvents(user, repo, issueNum, opts)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		allEvents = append(allEvents, events...)

		for _, e := range events {
			fmt.Println(e)
		}
		if !paginage || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage

	}
	return allEvents, nil
}
