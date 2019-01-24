package main

import (
	"encoding/json"
	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
)

func storeCommits(CommitMessages chan CommitEntry, BugChannel chan string, dbClient *elastic.Client) {
	for {
		ci := <- CommitMessages
		exists, err, oldEntry := Exists(dbClient, COMMIT_INDEX,  ci.Revision)
		if(err != nil){
			logrus.Info("Could not query commit from elastic: ", err)
			continue
		}
		if(exists){
			logrus.Info("Commit already in DB")
			data, err := oldEntry.Source.MarshalJSON()
			if err != nil {
				logrus.Info("Could not get old entry from elastic: %v", err)
			} else {
				err = json.Unmarshal(data,&ci)
				if err != nil {
					logrus.Warnf("Could unmarshall old entry from elastic: %v", err)
				}
			}
		}else{
			logrus.Info("Unknown Commit, storing in DB")
		}

		BugChannel <- "test";

		err = StoreElement(dbClient, COMMIT_INDEX, COMMIT_TYPE, ci, ci.Revision)
		if(err != nil){
			logrus.Info("Could not store commit to elastic: ", err)
			CommitMessages <- ci
			continue
		}

	}
}

