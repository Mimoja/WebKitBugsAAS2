package main

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func searchHandler(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Query not specified"))
		return
}

func main(){

	CommitMessages := make(chan CommitEntry)
	Bugs := make(chan string)


	logrus.Info("Connecting to DB")
	DBClient := DBConnect()
	logrus.Info("Starting SVN Fetching")
	go connectToSVN(CommitMessages)

	logrus.Info("Starting Commit Handler")
	go storeCommits(CommitMessages, Bugs,  DBClient)

	logrus.Info("Starting Bug Rater")
	go rateBugs(Bugs, DBClient)
	

	logrus.Info("Starting Webserver")
	http.HandleFunc("/", searchHandler)
	go http.ListenAndServe(":8080", nil)

	for {
		logrus.Info("Unhandled commits: %d Unhandled Bugs: %d", len(CommitMessages), len(Bugs))
		time.Sleep(1*time.Second)
	}

}
