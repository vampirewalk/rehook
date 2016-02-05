package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
)

const (
	// UserAgent header to use for outgoing HTTP requests
	UserAgent = "rehook/v0.1.1 (https://github.com/jstemmer/rehook)"
)

// flags
var (
	listenAddr = flag.String("http", ":9000", "Public HTTP listen address for incoming webhooks")
	adminAddr  = flag.String("admin", ":9001", "Private HTTP listen address for admin interface")
	database   = flag.String("db", "data.db", "Database file to use")
)

// Database constants
var (
	BucketHooks      = []byte("hooks")
	BucketComponents = []byte("components")
	BucketStats      = []byte("stats")
	BucketHookData   = []byte("hookdata")
	BucketIssue      = []byte("Issue")
	BucketEvent      = []byte("Event")
)

var (
	Interval = time.Second * 30
)

func main() {
	flag.Parse()

	// initialize database
	db, err := bolt.Open(*database, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("Could not open database: %s", err)
	}
	defer db.Close()

	if err := db.Update(initBuckets); err != nil {
		log.Fatal(err)
	}

	hookStore := &HookStore{db}

	// webhooks
	hh := &HookHandler{hookStore, db}
	router := httprouter.New()
	router.GET("/h/:id", hh.ReceiveHook)
	router.POST("/h/:id", hh.ReceiveHook)

	go func() {
		log.Printf("Listening on %s", *listenAddr)
		log.Print(http.ListenAndServe(*listenAddr, router))
	}()

	hooks, err := hookStore.List()
	if err != nil {
		log.Fatal("Fail to load hooks")
	}
	ist := NewIssueStatusTracker(NewIssueAPI(), hooks, db)

	// admin interface
	ah := &AdminHandler{hooks: hookStore, ist: ist}
	arouter := httprouter.New()
	arouter.Handler("GET", "/public/*path", http.StripPrefix("/public", http.FileServer(http.Dir("public"))))
	arouter.GET("/", ah.Index)
	arouter.Handler("GET", "/hooks", http.RedirectHandler("/", http.StatusMovedPermanently))

	arouter.GET("/hooks/new", ah.NewHook)
	arouter.POST("/hooks", ah.CreateHook)
	arouter.GET("/hooks/edit/:id", ah.EditHook)
	arouter.POST("/hooks/edit/:id", ah.UpdateHook)

	arouter.GET("/hooks/edit/:id/add", ah.AddComponent)
	arouter.POST("/hooks/edit/:id/create", ah.CreateComponent)
	arouter.GET("/hooks/edit/:id/edit/:c", ah.EditComponent)
	arouter.POST("/hooks/edit/:id/update/:c", ah.UpdateComponent)

	go func() {
		log.Printf("Admin interface on %s", *adminAddr)
		log.Print(http.ListenAndServe(*adminAddr, arouter))
	}()

	ist.RefreshIssues()
	ist.StartPolling()
}

func initBuckets(t *bolt.Tx) error {
	for _, name := range [][]byte{BucketHooks, BucketStats, BucketComponents} {
		if _, err := t.CreateBucketIfNotExists(name); err != nil {
			return err
		}
	}
	return nil
}
