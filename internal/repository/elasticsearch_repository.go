package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pkg/errors"
	"github.com/th1enq/es-demo/internal/domain"
	"go.uber.org/zap"
)

type elasticsearchRepository struct {
	client *elasticsearch.Client
	logger *zap.Logger
}

// NewElasticsearchRepository creates a new Elasticsearch repository
func NewElasticsearchRepository(client *elasticsearch.Client, logger *zap.Logger) domain.ElasticsearchRepository {
	return &elasticsearchRepository{
		client: client,
		logger: logger,
	}
}

// CreateIndex creates a new index in Elasticsearch
func (r *elasticsearchRepository) CreateIndex(indexName string) error {
	// Define index mapping for bank account projections
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"aggregateId": map[string]interface{}{
					"type": "keyword",
				},
				"email": map[string]interface{}{
					"type": "keyword",
				},
				"firstName": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"lastName": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"balance": map[string]interface{}{
					"properties": map[string]interface{}{
						"amount": map[string]interface{}{
							"type": "long",
						},
						"currency": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"createdAt": map[string]interface{}{
					"type": "date",
				},
				"updatedAt": map[string]interface{}{
					"type": "date",
				},
				"version": map[string]interface{}{
					"type": "long",
				},
				"totalDeposits": map[string]interface{}{
					"type": "long",
				},
				"totalWithdrawals": map[string]interface{}{
					"type": "long",
				},
				"transactionCount": map[string]interface{}{
					"type": "integer",
				},
				"status": map[string]interface{}{
					"type": "keyword",
				},
				"lastActivity": map[string]interface{}{
					"type": "date",
				},
			},
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
	}

	body, err := json.Marshal(mapping)
	if err != nil {
		r.logger.Error("Failed to marshal index mapping", zap.Error(err))
		return errors.Wrap(err, "json.Marshal")
	}

	req := esapi.IndicesCreateRequest{
		Index: indexName,
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(context.Background(), r.client)
	if err != nil {
		r.logger.Error("Failed to create index", zap.Error(err), zap.String("index", indexName))
		return errors.Wrap(err, "esapi.IndicesCreateRequest.Do")
	}
	defer res.Body.Close()

	if res.IsError() {
		// If index already exists, that's OK
		if strings.Contains(res.String(), "resource_already_exists_exception") {
			r.logger.Info("Index already exists", zap.String("index", indexName))
			return nil
		}
		r.logger.Error("Elasticsearch create index error", zap.String("response", res.String()))
		return errors.New(fmt.Sprintf("elasticsearch create index error: %s", res.String()))
	}

	r.logger.Info("Index created successfully", zap.String("index", indexName))
	return nil
}

// IndexDocument indexes a document in Elasticsearch
func (r *elasticsearchRepository) IndexDocument(indexName, documentID string, document interface{}) error {
	body, err := json.Marshal(document)
	if err != nil {
		r.logger.Error("Failed to marshal document", zap.Error(err))
		return errors.Wrap(err, "json.Marshal")
	}

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), r.client)
	if err != nil {
		r.logger.Error("Failed to index document", zap.Error(err), zap.String("index", indexName), zap.String("id", documentID))
		return errors.Wrap(err, "esapi.IndexRequest.Do")
	}
	defer res.Body.Close()

	if res.IsError() {
		r.logger.Error("Elasticsearch index document error", zap.String("response", res.String()))
		return errors.New(fmt.Sprintf("elasticsearch index document error: %s", res.String()))
	}

	r.logger.Debug("Document indexed successfully", zap.String("index", indexName), zap.String("id", documentID))
	return nil
}

// GetDocument retrieves a document from Elasticsearch
func (r *elasticsearchRepository) GetDocument(indexName, documentID string) (*domain.BankAccountElasticsearchProjection, error) {
	req := esapi.GetRequest{
		Index:      indexName,
		DocumentID: documentID,
	}

	res, err := req.Do(context.Background(), r.client)
	if err != nil {
		r.logger.Error("Failed to get document", zap.Error(err), zap.String("index", indexName), zap.String("id", documentID))
		return nil, errors.Wrap(err, "esapi.GetRequest.Do")
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, errors.New("document not found")
		}
		r.logger.Error("Elasticsearch get document error", zap.String("response", res.String()))
		return nil, errors.New(fmt.Sprintf("elasticsearch get document error: %s", res.String()))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		r.logger.Error("Failed to decode response", zap.Error(err))
		return nil, errors.Wrap(err, "json.Decode")
	}

	sourceData, ok := result["_source"]
	if !ok {
		return nil, errors.New("_source not found in response")
	}

	sourceBytes, err := json.Marshal(sourceData)
	if err != nil {
		r.logger.Error("Failed to marshal source data", zap.Error(err))
		return nil, errors.Wrap(err, "json.Marshal")
	}

	var projection domain.BankAccountElasticsearchProjection
	if err := json.Unmarshal(sourceBytes, &projection); err != nil {
		r.logger.Error("Failed to unmarshal to projection", zap.Error(err))
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return &projection, nil
}

// UpdateDocument updates a document in Elasticsearch
func (r *elasticsearchRepository) UpdateDocument(indexName, documentID string, document interface{}) error {
	updateDoc := map[string]interface{}{
		"doc": document,
	}

	body, err := json.Marshal(updateDoc)
	if err != nil {
		r.logger.Error("Failed to marshal update document", zap.Error(err))
		return errors.Wrap(err, "json.Marshal")
	}

	req := esapi.UpdateRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), r.client)
	if err != nil {
		r.logger.Error("Failed to update document", zap.Error(err), zap.String("index", indexName), zap.String("id", documentID))
		return errors.Wrap(err, "esapi.UpdateRequest.Do")
	}
	defer res.Body.Close()

	if res.IsError() {
		r.logger.Error("Elasticsearch update document error", zap.String("response", res.String()))
		return errors.New(fmt.Sprintf("elasticsearch update document error: %s", res.String()))
	}

	r.logger.Debug("Document updated successfully", zap.String("index", indexName), zap.String("id", documentID))
	return nil
}

// DeleteDocument deletes a document from Elasticsearch
func (r *elasticsearchRepository) DeleteDocument(indexName, documentID string) error {
	req := esapi.DeleteRequest{
		Index:      indexName,
		DocumentID: documentID,
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), r.client)
	if err != nil {
		r.logger.Error("Failed to delete document", zap.Error(err), zap.String("index", indexName), zap.String("id", documentID))
		return errors.Wrap(err, "esapi.DeleteRequest.Do")
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			r.logger.Warn("Document not found for deletion", zap.String("index", indexName), zap.String("id", documentID))
			return nil // Not an error if document doesn't exist
		}
		r.logger.Error("Elasticsearch delete document error", zap.String("response", res.String()))
		return errors.New(fmt.Sprintf("elasticsearch delete document error: %s", res.String()))
	}

	r.logger.Debug("Document deleted successfully", zap.String("index", indexName), zap.String("id", documentID))
	return nil
}

// Search searches documents in Elasticsearch
func (r *elasticsearchRepository) Search(indexName string, query map[string]interface{}) ([]*domain.BankAccountElasticsearchProjection, error) {
	body, err := json.Marshal(query)
	if err != nil {
		r.logger.Error("Failed to marshal search query", zap.Error(err))
		return nil, errors.Wrap(err, "json.Marshal")
	}

	req := esapi.SearchRequest{
		Index: []string{indexName},
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(context.Background(), r.client)
	if err != nil {
		r.logger.Error("Failed to search documents", zap.Error(err), zap.String("index", indexName))
		return nil, errors.Wrap(err, "esapi.SearchRequest.Do")
	}
	defer res.Body.Close()

	if res.IsError() {
		r.logger.Error("Elasticsearch search error", zap.String("response", res.String()))
		return nil, errors.New(fmt.Sprintf("elasticsearch search error: %s", res.String()))
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		r.logger.Error("Failed to decode search response", zap.Error(err))
		return nil, errors.Wrap(err, "json.Decode")
	}

	hits, ok := searchResult["hits"].(map[string]interface{})
	if !ok {
		return nil, errors.New("hits not found in search response")
	}

	hitsList, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, errors.New("hits array not found in search response")
	}

	var projections []*domain.BankAccountElasticsearchProjection
	for _, hit := range hitsList {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		sourceData, ok := hitMap["_source"]
		if !ok {
			continue
		}

		sourceBytes, err := json.Marshal(sourceData)
		if err != nil {
			r.logger.Error("Failed to marshal hit source", zap.Error(err))
			continue
		}

		var projection domain.BankAccountElasticsearchProjection
		if err := json.Unmarshal(sourceBytes, &projection); err != nil {
			r.logger.Error("Failed to unmarshal hit to projection", zap.Error(err))
			continue
		}

		projections = append(projections, &projection)
	}

	return projections, nil
}

// DeleteIndex deletes an index from Elasticsearch
func (r *elasticsearchRepository) DeleteIndex(indexName string) error {
	req := esapi.IndicesDeleteRequest{
		Index: []string{indexName},
	}

	res, err := req.Do(context.Background(), r.client)
	if err != nil {
		r.logger.Error("Failed to delete index", zap.Error(err), zap.String("index", indexName))
		return errors.Wrap(err, "esapi.IndicesDeleteRequest.Do")
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			r.logger.Warn("Index not found for deletion", zap.String("index", indexName))
			return nil // Not an error if index doesn't exist
		}
		r.logger.Error("Elasticsearch delete index error", zap.String("response", res.String()))
		return errors.New(fmt.Sprintf("elasticsearch delete index error: %s", res.String()))
	}

	r.logger.Info("Index deleted successfully", zap.String("index", indexName))
	return nil
}

// BulkIndex performs bulk indexing of documents
func (r *elasticsearchRepository) BulkIndex(indexName string, documents map[string]interface{}) error {
	var buf bytes.Buffer

	for docID, doc := range documents {
		// Add action and metadata
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    docID,
			},
		}
		actionBytes, err := json.Marshal(action)
		if err != nil {
			r.logger.Error("Failed to marshal bulk action", zap.Error(err))
			return errors.Wrap(err, "json.Marshal action")
		}
		buf.Write(actionBytes)
		buf.WriteByte('\n')

		// Add document
		docBytes, err := json.Marshal(doc)
		if err != nil {
			r.logger.Error("Failed to marshal bulk document", zap.Error(err))
			return errors.Wrap(err, "json.Marshal document")
		}
		buf.Write(docBytes)
		buf.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(context.Background(), r.client)
	if err != nil {
		r.logger.Error("Failed to perform bulk index", zap.Error(err), zap.String("index", indexName))
		return errors.Wrap(err, "esapi.BulkRequest.Do")
	}
	defer res.Body.Close()

	if res.IsError() {
		r.logger.Error("Elasticsearch bulk index error", zap.String("response", res.String()))
		return errors.New(fmt.Sprintf("elasticsearch bulk index error: %s", res.String()))
	}

	r.logger.Info("Bulk index completed successfully", zap.String("index", indexName), zap.Int("documents", len(documents)))
	return nil
}
