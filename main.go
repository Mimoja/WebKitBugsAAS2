package main

import (
	"github.com/Masterminds/vcs"
	"github.com/sirupsen/logrus"
)

func main(){

	CommitMessages := make(chan *vcs.CommitInfo)
	logrus.Info("Starting SVN Fetching")
	go connectToSVN(CommitMessages)

}
