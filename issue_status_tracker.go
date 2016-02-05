package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/google/go-github/github"
	"log"
	"net/http"
	"time"
)

type IssueStatusTracker struct {
	API    *IssueAPI
	Hooks  []Hook
	DB     *bolt.DB
	Ticker *time.Ticker // periodic ticker
	Add    chan Hook    // new URL channel
	Delete chan Hook    // new URL channel
}

func NewIssueStatusTracker(api *IssueAPI, hooks []Hook, db *bolt.DB) *IssueStatusTracker {
	return &IssueStatusTracker{API: NewIssueAPI(), Hooks: hooks, DB: db, Ticker: time.NewTicker(Interval), Add: make(chan Hook)}
}

func (i *IssueStatusTracker) RefreshIssues() error {
	return i.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(BucketIssue)
		if err != nil {
			return err
		}
		eb, err := tx.CreateBucketIfNotExists(BucketEvent)
		if err != nil {
			return err
		}
		hb := tx.Bucket(BucketHookData)
		if err != nil {
			return err
		}
		for _, hook := range i.Hooks {
			user := string(hb.Get([]byte(hook.ID + "_user")))
			repo := string(hb.Get([]byte(hook.ID + "_repo")))
			log.Printf("user: " + hook.User)
			log.Printf("repo: " + hook.Repo)
			issues, err := i.API.ListIssues(user, repo)
			if err != nil {
				return err
			}
			v, err := gobEncode(issues)
			if err != nil {
				return err
			}
			err = b.Put([]byte(hook.ID), v)
			if err != nil {
				return err
			}

			for _, issue := range issues {
				events, err := i.API.ListIssueEvents(user, repo, *issue.Number)
				if err != nil {
					return err
				}
				renameEvent := i.PickupRenameEvent(events)
				v, err := gobEncode(renameEvent)
				if err != nil {
					return err
				}
				err = eb.Put([]byte(hook.ID+string(*issue.Number)), v)
				if err != nil {
					return err
				}

			}
		}
		return nil
	})
}

func (i *IssueStatusTracker) PickupRenameEvent(events []github.IssueEvent) (e []github.IssueEvent) {
	for _, event := range events {
		if *event.Event == "renamed" {
			e = append(e, event)
		}
	}
	return e
}

func (i *IssueStatusTracker) StartPolling() {
	for {
		select {
		case <-i.Ticker.C:
			// When the ticker fires, it's time to harvest
			for _, h := range i.Hooks {
				err := i.trackStatus(h)
				if err != nil {
					log.Printf(err.Error())
				}
			}
		case u := <-i.Add:
			i.Hooks = append(i.Hooks, u)

		case u := <-i.Delete:
			targetIndex := -1
			for i, hook := range i.Hooks {
				if u.ID == hook.ID {
					targetIndex = i
					break
				}
			}
			i.Hooks = append(i.Hooks[:targetIndex], i.Hooks[targetIndex+1:]...)
		}

	}
}

func (i *IssueStatusTracker) trackStatus(hook Hook) error {
	return i.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(BucketIssue)
		if err != nil {
			return err
		}
		eb, err := tx.CreateBucketIfNotExists(BucketEvent)
		if err != nil {
			return err
		}
		hb := tx.Bucket(BucketHookData)
		if err != nil {
			return err
		}
		user := string(hb.Get([]byte(hook.ID + "_user")))
		repo := string(hb.Get([]byte(hook.ID + "_repo")))

		issues, err := i.API.ListIssues(user, repo)
		if err != nil {
			return err
		}
		v, err := gobEncode(issues)
		if err != nil {
			return err
		}
		err = b.Put([]byte(hook.ID), v)
		if err != nil {
			return err
		}

		for _, issue := range issues {
			newEvents, err := i.API.ListIssueEvents(user, repo, *issue.Number)
			if err != nil {
				return err
			}
			renameEvent := i.PickupRenameEvent(newEvents)
			log.Printf("%d new rename event", len(renameEvent))
			//load events from db)
			v := eb.Get([]byte(hook.ID + string(*issue.Number)))
			if v != nil {
				oldRenameEvents := make([]github.IssueEvent, 0)
				err = gobDecode(v, &oldRenameEvents)
				if err != nil {
					return err
				}
				log.Printf("%d old rename event", len(oldRenameEvents))
				if len(renameEvent) > len(oldRenameEvents) {
					//call hook
					log.Printf("call")
					i.invokeHook(hook.ID, user, repo, issue)
				}
			} else {
				if len(renameEvent) > 0 {
					//call hook
					log.Printf("call")
					i.invokeHook(hook.ID, user, repo, issue)
				}
			}

			//save in db
			v, err = gobEncode(renameEvent)
			if err != nil {
				return err
			}

			err = eb.Put([]byte(hook.ID+string(*issue.Number)), v)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (i *IssueStatusTracker) invokeHook(hookID, user, repoName string, issue github.Issue) error {
	client := github.NewClient(nil)
	repo, _, err := client.Repositories.Get(user, repoName)
	if err != nil {
		return err
	}

	action := "updated"
	event := github.IssueActivityEvent{Action: &action, Issue: &issue, Repo: repo, Sender: issue.User}
	b, err := json.Marshal(event)
	if err != nil {
		log.Printf("Struct to JSON error")
	}
	url := "http://localhost" + *listenAddr + "/h/" + hookID
	log.Printf("POST JSON to %s", url)
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request forward error: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("request forward unexpected status code received: %d", resp.StatusCode)
	}

	return nil
}

func (i *IssueStatusTracker) AddHook(h *Hook) {
	// Adding a new URL is as simple as tossing it onto a channel.
	i.Add <- *h
}

func (i *IssueStatusTracker) DeleteHook(h *Hook) {
	// Adding a new URL is as simple as tossing it onto a channel.
	i.Delete <- *h
}
