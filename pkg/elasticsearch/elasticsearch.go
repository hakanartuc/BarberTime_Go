package elasticsearch

import (
	"barbertime_go/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

var Client *elasticsearch.Client

func Connect() error {
	cfg := elasticsearch.Config{
		Addresses: []string{config.AppConfig.ElasticsearchURL},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Bağlantıyı test et
	res, err := es.Info()
	if err != nil {
		return fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	Client = es
	log.Println("Elasticsearch connected successfully")
	return nil
}

func IndexDocument(index string, id string, document interface{}) error {
	body, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       strings.NewReader(string(body)),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), Client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch error: %s", res.String())
	}

	return nil
}

func Search(index string, query map[string]interface{}, from, size int) (map[string]interface{}, error) {
	queryBody, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	res, err := Client.Search(
		Client.Search.WithContext(context.Background()),
		Client.Search.WithIndex(index),
		Client.Search.WithBody(strings.NewReader(string(queryBody))),
		Client.Search.WithFrom(from),
		Client.Search.WithSize(size),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

func DeleteDocument(index, id string) error {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: id,
	}

	res, err := req.Do(context.Background(), Client)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("Elasticsearch error: %s", res.String())
	}

	return nil
}

func CreateIndex(index string, mapping map[string]interface{}) error {
	mappingBody, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  strings.NewReader(string(mappingBody)),
	}

	res, err := req.Do(context.Background(), Client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// Index zaten varsa hata verme
		if res.StatusCode == 400 {
			return nil
		}
		return fmt.Errorf("Elasticsearch error: %s", res.String())
	}

	return nil
}
