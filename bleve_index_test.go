package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBleveIndex(t *testing.T) {
	// Create a temporary directory for the index
	tempDir, err := os.MkdirTemp("", "bleve_test_index")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize the index
	index, err := NewBleveIndex(tempDir)
	if err != nil {
		t.Fatalf("Failed to create Bleve index: %v", err)
	}
	defer index.Close()

	// Test data
	entities := []Entity{
		{
			ID:         "entity1",
			SourceID:   "source1",
			ProjectIDs: []string{"project1", "project2"},
		},
		{
			ID:         "entity2",
			SourceID:   "source2",
			ProjectIDs: []string{"project2", "project3"},
		},
		{
			ID:         "entity3",
			SourceID:   "source3",
			ProjectIDs: []string{"project4"},
		},
	}

	// Index the entities
	for _, entity := range entities {
		if err := index.IndexEntity(entity); err != nil {
			t.Fatalf("Failed to index entity %s: %v", entity.ID, err)
		}
	}

	// Test 1: Search for entities with project1
	t.Run("SearchByProjectID_Project1", func(t *testing.T) {
		results, err := index.SearchByProjectID("project1")
		if err != nil {
			t.Fatalf("Failed to search by project ID: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result for project1, got %d", len(results))
		}

		if len(results) > 0 && results[0].ID != "entity1" {
			t.Errorf("Expected entity1, got %s", results[0].ID)
		}
	})

	// Test 2: Search for entities with project2 (should return 2 entities)
	t.Run("SearchByProjectID_Project2", func(t *testing.T) {
		results, err := index.SearchByProjectID("project2")
		if err != nil {
			t.Fatalf("Failed to search by project ID: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results for project2, got %d", len(results))
		}

		// Check if both entity1 and entity2 are in the results
		foundEntity1 := false
		foundEntity2 := false
		for _, result := range results {
			if result.ID == "entity1" {
				foundEntity1 = true
			}
			if result.ID == "entity2" {
				foundEntity2 = true
			}
		}

		if !foundEntity1 {
			t.Error("Expected to find entity1 in results for project2")
		}
		if !foundEntity2 {
			t.Error("Expected to find entity2 in results for project2")
		}
	})

	// Test 3: Get ID and SourceID by project ID
	t.Run("GetIDAndSourceIDByProjectID_Project2", func(t *testing.T) {
		results, err := index.GetIDAndSourceIDByProjectID("project2")
		if err != nil {
			t.Fatalf("Failed to get ID and SourceID by project ID: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results for project2, got %d", len(results))
		}

		// Check if the results contain the expected IDs and SourceIDs
		foundEntity1 := false
		foundEntity2 := false
		for _, result := range results {
			if result["id"] == "entity1" && result["source_id"] == "source1" {
				foundEntity1 = true
			}
			if result["id"] == "entity2" && result["source_id"] == "source2" {
				foundEntity2 = true
			}
		}

		if !foundEntity1 {
			t.Error("Expected to find entity1 with source1 in results for project2")
		}
		if !foundEntity2 {
			t.Error("Expected to find entity2 with source2 in results for project2")
		}
	})

	// Test 4: Search for non-existent project
	t.Run("SearchByProjectID_NonExistent", func(t *testing.T) {
		results, err := index.SearchByProjectID("nonexistent")
		if err != nil {
			t.Fatalf("Failed to search by project ID: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 results for nonexistent project, got %d", len(results))
		}
	})
}

func TestDefaultIndexPath(t *testing.T) {
	expectedPath := filepath.Join(".", "bleve_index")
	actualPath := DefaultIndexPath()

	if actualPath != expectedPath {
		t.Errorf("Expected default index path %s, got %s", expectedPath, actualPath)
	}
}
