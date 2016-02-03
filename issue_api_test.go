package main

import "testing"

func TestListIssue(t *testing.T) {
	tracker := IssueAPI{User: "vampirewalk", Repo: "hooksim"}
	tracker.ListIssues()
}

func TestListIssueEvent(t *testing.T) {
	tracker := IssueAPI{User: "vampirewalk", Repo: "hooksim"}
	tracker.ListIssueEvents(1)
}
