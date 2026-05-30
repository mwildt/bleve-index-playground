package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
)

// Entity represents the data structure to be indexed
type Entity struct {
	ID         string   `json:"id"`
	SourceID   string   `json:"source_id"`
	ProjectIDs []string `json:"project_ids"`
}

// BleveIndex manages the Bleve index for entities
type BleveIndex struct {
	index bleve.Index
}

// NewBleveIndex creates a new Bleve index instance
func NewBleveIndex(indexPath string) (*BleveIndex, error) {
	// Create the index directory if it doesn't exist
	if err := os.MkdirAll(indexPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create index directory: %v", err)
	}

	// Define a custom index mapping for Entity
	entityMapping := bleve.NewDocumentMapping()
	entityMapping.AddFieldMappingsAt("id", bleve.NewTextFieldMapping())           // Store and index ID
	entityMapping.AddFieldMappingsAt("source_id", bleve.NewTextFieldMapping())   // Store and index SourceID
	entityMapping.AddFieldMappingsAt("project_ids", bleve.NewTextFieldMapping()) // Store and index ProjectIDs

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("entity", entityMapping)

	// Open or create the index
	index, err := bleve.Open(indexPath)
	if err != nil {
		// If the index doesn't exist, create a new one
		index, err = bleve.New(indexPath, indexMapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create new index: %v", err)
		}
	}

	return &BleveIndex{index: index}, nil
}

// IndexEntity adds or updates an entity in the index
func (bi *BleveIndex) IndexEntity(entity Entity) error {
	// Marshal the entity to JSON bytes for storage
	data, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %v", err)
	}
	// Index the JSON data as a document
	return bi.index.Index(entity.ID, data)
}

// SearchByProjectID searches for entities containing the given project ID
// Returns a list of entities matching the project ID
func (bi *BleveIndex) SearchByProjectID(projectID string) ([]Entity, error) {
	// Create a term query to search for the project ID in the project_ids field
	query := bleve.NewTermQuery(projectID, "project_ids")
	searchRequest := bleve.NewSearchRequest(query)

	// Execute the search
	searchResults, err := bi.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search index: %v", err)
	}

	// Extract the entities from the search results
	var entities []Entity
	for _, hit := range searchResults.Hits {
		// Retrieve the raw document data
		docBytes, err := bi.index.GetInternal([]byte(hit.ID))
		if err != nil {
			log.Printf("Warning: failed to retrieve document for ID %s: %v", hit.ID, err)
			continue
		}
		// Unmarshal the JSON data into an Entity
		var entity Entity
		if err := json.Unmarshal(docBytes, &entity); err != nil {
			log.Printf("Warning: failed to unmarshal document for ID %s: %v", hit.ID, err)
			continue
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

// GetIDAndSourceIDByProjectID returns the ID and SourceID for entities matching the project ID
func (bi *BleveIndex) GetIDAndSourceIDByProjectID(projectID string) ([]map[string]string, error) {
	entities, err := bi.SearchByProjectID(projectID)
	if err != nil {
		return nil, err
	}

	// Extract ID and SourceID from each entity
	var results []map[string]string
	for _, entity := range entities {
		results = append(results, map[string]string{
			"id":       entity.ID,
			"source_id": entity.SourceID,
		})
	}

	return results, nil
}

// Close closes the Bleve index
func (bi *BleveIndex) Close() error {
	return bi.index.Close()
}

// DefaultIndexPath returns the default path for the Bleve index
func DefaultIndexPath() string {
	return filepath.Join(".", "bleve_index")
}
