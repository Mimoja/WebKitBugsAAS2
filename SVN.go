package main

import (
	"encoding/xml"
	"fmt"
	"github.com/Masterminds/vcs"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

func connectToSVN(CommitMessages chan CommitEntry) {

	remote := "https://svn.webkit.org/repository/webkit/trunk"
	local := "./webkit-svn"

	if _, err := os.Stat(local); os.IsNotExist(err) {
		os.Mkdir(local, os.ModePerm)
	}

	logrus.Info("Creating SVN repo")
	repo, err := vcs.NewSvnRepo(remote, local)

	if err != nil {
		logrus.Error("Could not connevt to repo: ", err)
		return
	}

	err = repo.Get()
	if err != nil {
		logrus.Errorf("Unable to checkout SVN repo. Err was %v", err)
		return
	}

	lastKnownVersion := 1

	for {
		logrus.Info("Updating SVN repo")
		err = repo.Update()
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

		latestVersion, err := strconv.Atoi(v)
		if err != nil {
			logrus.Errorf("Unable to convert SVN version. Err was %s", err)
			return
		}
		logrus.Info("Latest version is: ", latestVersion)

		paging := 5000
		for lastKnownVersion < latestVersion {

			if latestVersion-lastKnownVersion < paging {
				paging = latestVersion - lastKnownVersion
			}

			logrus.Infof("Getting CommitInfo for %d:%d", lastKnownVersion, lastKnownVersion+paging)
			cis, err := getCommitInfos(repo, lastKnownVersion, lastKnownVersion+paging)
			if err != nil {
				if err.Error() == "Revision unavailable" {
					logrus.Error("Revision unavailable.Help!")
					continue
				}
				logrus.Errorf("Unable to get svn commit message. Err was: %s", err)
				continue
			}

			for _, ci := range cis {
				CommitMessages <- ci
			}
			lastKnownVersion += paging
		}
		timer := time.NewTimer(10 * time.Minute)
		<-timer.C
	}
}

func getCommitInfos(s *vcs.SvnRepo, from int, to int) ([]CommitEntry, error) {

	out, err := s.RunFromDir("svn", "log", "-r", fmt.Sprintf("%d:%d", from, to), "--xml")
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve commit information: %v", err)
	}

	type Logentry struct {
		Author string `xml:"author"`
		Date   string `xml:"date"`
		Msg    string `xml:"msg"`
	}
	type Log struct {
		XMLName xml.Name   `xml:"log"`
		Logs    []Logentry `xml:"logentry"`
	}

	logs := &Log{}
	err = xml.Unmarshal(out, &logs)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshall commit information %v", err)
	}
	if len(logs.Logs) == 0 {
		return nil, fmt.Errorf("Revision unavailable")
	}
	var cis []CommitEntry
	for i, log := range logs.Logs {
		ci := CommitEntry{
			Revision: strconv.Itoa(from + i),
			CommitInfo: CommitInfo{
				Author:  log.Author,
				Message: log.Msg,
			},
		}
		if len(log.Date) > 0 {
			ci.CommitInfo.Date, err = time.Parse(time.RFC3339Nano, log.Date)
			if err != nil {
				return nil, fmt.Errorf("Unable to retrieve commit information: %v", err)
			}
			cis = append(cis, ci)
		}
	}
	return cis, nil
}
