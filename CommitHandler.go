package main

import (
	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
	"math"
	"net/url"
	"regexp"
	"strconv"
)

func storeCommits(CommitMessages chan CommitEntry, BugChannel chan string, dbClient *elastic.Client) {
	for {
		ci := <-CommitMessages
		messageLength := float64(len(ci.CommitInfo.Message))
		logrus.Infof("Commit %d message is: %s", ci.Revision, ci.CommitInfo.Message[:int(math.Min(messageLength, 30))]+"...")
		exists, err, _ := Exists(dbClient, COMMIT_INDEX, strconv.Itoa(ci.Revision))
		if err != nil {
			logrus.Info("Could not query commit %s from elastic: ", err, ci.Revision)
			continue
		}
		if exists {
			logrus.Infof("Commit %d already in DB", ci.Revision)
			continue
		} else {
			logrus.Info("Unknown Commit, storing in DB")
		}

		bugTrackerUrl := `(bugs.webkit.org|bugzilla.opendarwin.org)/show_bug\.cgi\?id=[1-9]*`

		reg, _ := regexp.Compile(bugTrackerUrl)

		for _, bugURL := range reg.FindAllString(ci.CommitInfo.Message, -1) {
			logrus.Info("Found Bug URL: ", bugURL)
			parsedURL, _ := url.Parse(bugURL)
			bugID := parsedURL.Query().Get("id")
			BugChannel <- bugID
			ci.Bugs = append(ci.Bugs, bugID)
		}

		err = StoreElement(dbClient, COMMIT_INDEX, COMMIT_TYPE, ci, strconv.Itoa(ci.Revision))
		if err != nil {
			logrus.Error("Could not store commit to elastic: ", err)
			CommitMessages <- ci
			continue
		}

	}
}
