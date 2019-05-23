package main

import (
	"fmt"
	"github.com/Masterminds/vcs"
	"github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func connectToGIT(CommitMessages chan CommitEntry) {

	remote := "https://github.com/webkit/webkit"
	local := "../webkit"

	logrus.Info("Creating GIT repo")
	repo, err := vcs.NewGitRepo(remote, local)

	if err != nil {
		logrus.Error("Could not connevt to repo: ", err)
		return
	}

	if _, err := os.Stat(local); os.IsNotExist(err) {
		err = repo.Get()
		if err != nil {
			logrus.Errorf("Unable to checkout GIT repo. Err was %v", err)
			return
		}
	}


	for {
		logrus.Info("Updating GIT repo")
		//err = repo.Update()
		if err != nil {
			logrus.Errorf("Unable update GIT repo. Err was %s", err)
			return
		}

		logrus.Info("Getting GIT latest version")
		v, err := repo.Version()
		if err != nil {
			logrus.Warnf("Unable to get current GIT version. Err was %s", err)
			return
		}

		logrus.Info("Latest version is: ", v, repo.Vcs())

		commits, err := getLogs(repo, "%H")
		if err != nil {
			logrus.Errorf("Unable to retrieve commit information: %v", err)
			continue
		}

		authors, err := getLogs(repo, "%an")
		if err != nil {
			logrus.Errorf("Unable to retrieve author information: %v", err)
			continue
		}

		dates, err := getLogs(repo, "%aD")
		if err != nil {
			logrus.Errorf("Unable to retrieve author information: %v", err)
			continue
		}

		bodies, err := getLogs(repo, "%B")
		if err != nil {
			logrus.Errorf("Unable to retrieve author information: %v", err)
			continue
		}

		entryNumber := len(commits)
		if len(authors) != entryNumber || len(dates) != entryNumber || len(bodies) != entryNumber {
			logrus.Error("Something went horrible wrong. Wrong count of commit components")
			continue
		}
		re := regexp.MustCompile("webkit/trunk@(.*?)\\ ")

		for i := 0; i < entryNumber-1; i++ {
			info := CommitInfo{
				CommitID: strings.TrimSpace(commits[i]),
				Author:   strings.TrimSpace(authors[i]),
				Date:     strings.TrimSpace(dates[i]),
				Message:  strings.TrimSpace(bodies[i]),
			}
			match := re.FindStringSubmatch(info.Message)
			revision, err := strconv.Atoi(match[1])

			if err != nil {
				logrus.Error("Could not extract revision")
				continue
			}

			CommitMessages <- CommitEntry{
				Revision:   revision,
				CommitInfo: info,
			}
		}

		timer := time.NewTimer(10 * time.Minute)
		<-timer.C
	}
}

func getLogs(s *vcs.GitRepo, format string) ([]string, error) {
	bytess, e := s.RunFromDir("git", "log", fmt.Sprintf(`--pretty=format:%s___________________---------MIMOJA--------___________________`, format))

	logs := string(bytess)
	return strings.Split(logs, "___________________---------MIMOJA--------___________________"), e
}
