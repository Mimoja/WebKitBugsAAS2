package main

import (
	"context"
	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
	"time"
)


type BugVisibility string

const PUBLIC = BugVisibility("public")
const PRIVATE = BugVisibility("private")
const UNKNOWN = BugVisibility("unknown")
const IGNORED = BugVisibility("ignored")

const COMMIT_INDEX = "commits"
const BUG_INDEX = "bugs"
const COMMIT_TYPE = "commit"
const BUG_TYPE = "bug"


type CommitEntry struct {
	Revision string
	CommitInfo CommitInfo
	Bugs []string
}

type BugEntry struct{
	ID string
	Visibility BugVisibility
}

type CommitInfo struct {
	Author string
	Date time.Time
	Message string
}

func DBConnect() (*elastic.Client){
	client, err := elastic.NewClient()
	if err != nil {
		logrus.WithError(err).Panic("Could not create elastic client")
	}

	// Getting the ES version number is quite common, so there's a shortcut
	esversion, err := client.ElasticsearchVersion("http://127.0.0.1:9200")
	if err != nil {
		logrus.WithError(err).Panic("Could not connect to elastic")
	}
	logrus.Infof("Elasticsearch version %s", esversion)
	return client
}


func updateElement(client *elastic.Client,index string, typeString string,  id string, field string, entry interface{},) {
	_, err := client.Update().
		Index(index).
		Type(typeString).
		Id(id).
		Doc(map[string]interface{}{field: entry}).
		Do(context.Background())

	if err != nil {
		logrus.Errorf("Error while updating %s", entry)
	}
	logrus.Infof("updated %s", entry)
}

func StoreElement(client *elastic.Client, index string, typeString string, entry interface{}, id string) error {

	is := client.Index().BodyJson(entry)
	return store(is, index, typeString, id)

}


func store(is *elastic.IndexService, index string, typeString string, id string) error {
	is = is.Index(index).Type(typeString).Id(id)

	put1, err := is.Do(context.Background())
	if err != nil {
		// Handle error
		logrus.WithError(err).Error("Could not execute elastic search")
		return err
	}
	logrus.WithField("dbentry", put1).Infof("Indexed %s to index %s, type %s", put1.Id, put1.Index, put1.Type)
	return nil
}

func Exists(client *elastic.Client,index string, id string) (bool, error, *elastic.GetResult) {
	get, err := client.Get().
		Index(index).
		Id(id).
		Do(context.Background())

	if err != nil {
		switch {
		case elastic.IsNotFound(err):
			return false, nil, get
		case elastic.IsTimeout(err):
			logrus.WithError(err).Errorf("Timeout retrieving document: %v", err)
			return false, err, get
		case elastic.IsConnErr(err):
			logrus.WithError(err).Errorf("Connection problem: %v", err)
			return false, err, get
		default:
			logrus.WithError(err).Errorf("Unknown error: %v", err)
			return false, err, get
		}
	}
	return true, nil, get
}


