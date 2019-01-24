package main

import (
	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
	"net/http"
)

const bugsWebkitOrg = "https://bugs.webkit.org/show_bug.cgi?id="

func rateBugs(Bugs chan string, dbClient *elastic.Client) {
	for {
		bugID := <-Bugs
		logrus.Info("Found Bug ID: ", bugID)

		exists, err, _ := Exists(dbClient, BUG_INDEX, bugID)
		if err != nil {
			logrus.Info("Could not query commit %s from elastic: ", err, bugID)
		}
		if err == nil && exists {
			logrus.Infof("Bug %s already in DB", bugID)
			continue
		}

		bugEntry := BugEntry{
			ID:         bugID,
			Visibility: UNKNOWN,
		}
		logrus.Info("Rating Bug: ", bugID)
		response, err := http.Get(bugsWebkitOrg + bugID)
		if err != nil {
			if response != nil && response.StatusCode == 401 {
				bugEntry.Visibility = PRIVATE
			} else {
				logrus.Error("Could not fetch bugtracker: ", err)

			}
		} else {
			defer response.Body.Close()
			if response != nil && response.StatusCode == 401 {
				bugEntry.Visibility = PRIVATE
			} else {
				bugEntry.Visibility = PUBLIC
			}
		}
		logrus.Infof("Bug %s is : %s", bugID, bugEntry.Visibility)

		err = StoreElement(dbClient, BUG_INDEX, BUG_TYPE, bugEntry, bugID)
		if err != nil {
			logrus.Error("Could not store bug to elastic: ", err)
			continue
		}
	}
}
