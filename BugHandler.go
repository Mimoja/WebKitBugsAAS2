package main

import "github.com/olivere/elastic"

func rateBugs(Bugs chan string, dbClient *elastic.Client) {
	for {
		<-Bugs
	}
}
