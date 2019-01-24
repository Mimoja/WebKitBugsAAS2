package main

import (
	"github.com/sirupsen/logrus"
)

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
	select {}
}
