package main

import (
	"context"
	"dataLoad/pkg/GetHostInfo"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
)

func create(host_info *GetHostInfo.Host_info_type, esClient *elastic.Client) error {
	host_info_json, err := json.Marshal(host_info)
	if err != nil {
		return err
	}
	put, err := esClient.Index().Index("info").Type("hosts").Index("1").BodyJson(host_info_json).Do(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("indexed %d to index %s, type %s \n", put.Id, put.Index, put.Type)
	return nil
}

func main() {
	host_info := GetHostInfo.Get_host_info()
	fmt.Printf("host_info has been got!")
	esHost := "http://192.168.0.178:9200"
	fmt.Printf("init a esClient.")
	esClient, err := elastic.NewClient(elastic.SetURL(esHost), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}

	err = create(&host_info, esClient)
	if err != nil {
		panic(err)
	}

}
