package main

import (
	"fmt"
	"github.com/Masterminds/vcs"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

func connectToSVN(CommitMessages chan CommitEntry){

	remote := "https://svn.webkit.org/repository/webkit/trunk"
	local := "/home/mimoja/webkit-svn"

	if _, err := os.Stat(local); os.IsNotExist(err) {
		os.Mkdir(local, os.ModePerm)
	}

	logrus.Info("Creating SVN repo")
	repo, err := vcs.NewSvnRepo(remote, local)

	if(err != nil){
		logrus.Error("Could not connevt to repo: ", err)
		return
	}

	//err = repo.Get()
	if err != nil {
		logrus.Errorf("Unable to checkout SVN repo. Err was %s", err)
		return
	}


	lastKnownVersion := 1;

	for {
		logrus.Info("Updating SVN repo")
		//err = repo.Update()
		if err != nil {
			logrus.Errorf("Unable update SVN repo. Err was %s", err)
			return
		}

		logrus.Info("Getting SVN latest version")
		v, err := repo.Version()
		if err != nil {
			logrus.Warnf("Unable to get current SVN version. Err was %s", err)
			return
		}

		latestVersion, err := strconv.Atoi(v);
		if err != nil {
			logrus.Errorf("Unable to convert SVN version. Err was %s", err)
			return
		}
		logrus.Info("Latest version is: ", latestVersion)

		for ; lastKnownVersion <= latestVersion; lastKnownVersion++ {
			lastCommit := fmt.Sprint(lastKnownVersion)

			logrus.Info("Getting CommitInfo for ", lastCommit)
			ci, err := repo.CommitInfo(lastCommit)

			if err != nil {
				if err.Error() == "Revision unavailable" {
					logrus.Info("Skipping non existing Revision: ", lastCommit)
					continue
				}
				logrus.Errorf("Unable to svn commit message. Err was %s", err)
				continue
			}

			CommitMessages <- CommitEntry{
				Revision: ci.Commit,
				CommitInfo: CommitInfo{
					Author:  ci.Author,
					Date:    ci.Date,
					Message: ci.Message,
				},
			}

		}
		timer := time.NewTimer(10 * time.Minute)
		<-timer.C
	}
}