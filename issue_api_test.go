package main

import "testing"

func TestListIssue(t *testing.T) {
	tracker := NewIssueAPI()
	tracker.ListIssues("vampirewalk", "hooksim")
}

func TestListIssueEvent(t *testing.T) {
	tracker := NewIssueAPI()
	tracker.ListIssueEvents("vampirewalk", "hooksim", 1)
}
