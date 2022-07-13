package main

import (
	"context"
	"dataLoad/pkg/GetHostInfo"
	"fmt"
	"github.com/olivere/elastic/v7"
	"log"
	"os"
)

var esClient *elastic.Client
var esHost = "http://192.168.0.178:9200"

func init() {
	fmt.Printf("begin to init es client\n")
	errorlog := log.New(os.Stdout, "app", log.LstdFlags)

	var err error
	esClient, err = elastic.NewClient(elastic.SetErrorLog(errorlog), elastic.SetURL(esHost), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}
	info, code, err := esClient.Ping(esHost).Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("ES return with code %d and version %s \n", code, info.Version.Number)
	esversionCode, err := esClient.ElasticsearchVersion(esHost)
	if err != nil {
		panic(err)
	}
	fmt.Printf("es version %s\n", esversionCode)
}

func create(hostInfo *GetHostInfo.Host_info_type) {
	put, err := esClient.Index().Index("info").Type("hosts").Index("1").BodyJson(hostInfo).Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("indexed %d to index %s, type %s \n", put.Id, put.Index, put.Type)
}

func main() {
	hostInfo := GetHostInfo.Get_host_info()
	create(&hostInfo)
}
