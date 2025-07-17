package storage_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/exler/fileigloo/storage"
)

func TestNewLocalStorage(t *testing.T) {
	t.Run("creates storage with trailing slash", func(t *testing.T) {
		tempDir := t.TempDir()
		s, err := storage.NewLocalStorage(tempDir)
		if err != nil {
			t.Fatalf("Failed to create local storage: %v", err)
		}
		if s.Type() != "local" {
			t.Errorf("Expected storage type 'local', got '%s'", s.Type())
		}
	})

	t.Run("adds trailing slash if missing", func(t *testing.T) {
		tempDir := strings.TrimSuffix(t.TempDir(), "/")
		s, err := storage.NewLocalStorage(tempDir)
		if err != nil {
			t.Fatalf("Failed to create local storage: %v", err)
		}
		if s.Type() != "local" {
			t.Errorf("Expected storage type 'local', got '%s'", s.Type())
		}
	})
}

func TestLocalStorage_Put(t *testing.T) {
	tempDir := t.TempDir()
	s, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	t.Run("puts file with metadata", func(t *testing.T) {
		ctx := context.Background()
		filename := "test.txt"
		content := "Hello, World!"
		reader := bytes.NewBufferString(content)

		metadata := storage.Metadata{
			Filename:      "original.txt",
			ContentType:   "text/plain",
			ContentLength: "13",
			PasswordHash:  "hash123",
			ExpiresAt:     "2025-12-31T23:59:59Z",
		}

		err := s.Put(ctx, filename, reader, metadata)
		if err != nil {
			t.Fatalf("Failed to put file: %v", err)
		}

		// Verify file exists
		filePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("File was not created: %s", filePath)
		}

		// Verify metadata file exists
		metadataPath := filePath + ".metadata"
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			t.Errorf("Metadata file was not created: %s", metadataPath)
		}

		// Verify file content
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(fileContent) != content {
			t.Errorf("File content mismatch. Expected: %s, Got: %s", content, string(fileContent))
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		ctx := context.Background()
		filename := "overwrite.txt"

		// First write
		content1 := "First content"
		reader1 := bytes.NewBufferString(content1)
		metadata1 := storage.Metadata{
			Filename:      "first.txt",
			ContentType:   "text/plain",
			ContentLength: "13",
		}

		err := s.Put(ctx, filename, reader1, metadata1)
		if err != nil {
			t.Fatalf("Failed to put first file: %v", err)
		}

		// Second write (overwrite)
		content2 := "Second content"
		reader2 := bytes.NewBufferString(content2)
		metadata2 := storage.Metadata{
			Filename:      "second.txt",
			ContentType:   "text/plain",
			ContentLength: "14",
		}

		err = s.Put(ctx, filename, reader2, metadata2)
		if err != nil {
			t.Fatalf("Failed to put second file: %v", err)
		}

		// Verify content was overwritten
		filePath := filepath.Join(tempDir, filename)
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(fileContent) != content2 {
			t.Errorf("File content was not overwritten. Expected: %s, Got: %s", content2, string(fileContent))
		}
	})
}

func TestLocalStorage_Get(t *testing.T) {
	tempDir := t.TempDir()
	s, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	t.Run("gets existing file", func(t *testing.T) {
		ctx := context.Background()
		filename := "test.txt"
		content := "Hello, World!"

		// Put file first
		reader := bytes.NewBufferString(content)
		metadata := storage.Metadata{
			Filename:      "original.txt",
			ContentType:   "text/plain",
			ContentLength: "13",
		}

		err := s.Put(ctx, filename, reader, metadata)
		if err != nil {
			t.Fatalf("Failed to put file: %v", err)
		}

		// Get file
		fileReader, err := s.Get(ctx, filename)
		if err != nil {
			t.Fatalf("Failed to get file: %v", err)
		}
		defer fileReader.Close()

		// Verify content
		gotContent, err := io.ReadAll(fileReader)
		if err != nil {
			t.Fatalf("Failed to read file content: %v", err)
		}
		if string(gotContent) != content {
			t.Errorf("File content mismatch. Expected: %s, Got: %s", content, string(gotContent))
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		ctx := context.Background()
		filename := "non-existent.txt"

		_, err := s.Get(ctx, filename)
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
		if !s.FileNotExists(err) {
			t.Errorf("Expected FileNotExists to return true for error: %v", err)
		}
	})
}

func TestLocalStorage_GetWithMetadata(t *testing.T) {
	tempDir := t.TempDir()
	s, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	t.Run("gets file with metadata", func(t *testing.T) {
		ctx := context.Background()
		filename := "test.txt"
		content := "Hello, World!"

		// Put file first
		reader := bytes.NewBufferString(content)
		originalMetadata := storage.Metadata{
			Filename:      "original.txt",
			ContentType:   "text/plain",
			ContentLength: "13",
			PasswordHash:  "hash123",
			ExpiresAt:     "2025-12-31T23:59:59Z",
		}

		err := s.Put(ctx, filename, reader, originalMetadata)
		if err != nil {
			t.Fatalf("Failed to put file: %v", err)
		}

		// Get file with metadata
		fileReader, metadata, err := s.GetWithMetadata(ctx, filename)
		if err != nil {
			t.Fatalf("Failed to get file with metadata: %v", err)
		}
		defer fileReader.Close()

		// Verify content
		gotContent, err := io.ReadAll(fileReader)
		if err != nil {
			t.Fatalf("Failed to read file content: %v", err)
		}
		if string(gotContent) != content {
			t.Errorf("File content mismatch. Expected: %s, Got: %s", content, string(gotContent))
		}

		// Verify metadata
		if metadata.Filename != originalMetadata.Filename {
			t.Errorf("Filename mismatch. Expected: %s, Got: %s", originalMetadata.Filename, metadata.Filename)
		}
		if metadata.ContentType != originalMetadata.ContentType {
			t.Errorf("ContentType mismatch. Expected: %s, Got: %s", originalMetadata.ContentType, metadata.ContentType)
		}
		if metadata.ContentLength != originalMetadata.ContentLength {
			t.Errorf("ContentLength mismatch. Expected: %s, Got: %s", originalMetadata.ContentLength, metadata.ContentLength)
		}
		if metadata.PasswordHash != originalMetadata.PasswordHash {
			t.Errorf("PasswordHash mismatch. Expected: %s, Got: %s", originalMetadata.PasswordHash, metadata.PasswordHash)
		}
		if metadata.ExpiresAt != originalMetadata.ExpiresAt {
			t.Errorf("ExpiresAt mismatch. Expected: %s, Got: %s", originalMetadata.ExpiresAt, metadata.ExpiresAt)
		}
	})
}

func TestLocalStorage_GetOnlyMetadata(t *testing.T) {
	tempDir := t.TempDir()
	s, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	t.Run("gets only metadata", func(t *testing.T) {
		ctx := context.Background()
		filename := "test.txt"
		content := "Hello, World!"

		// Put file first
		reader := bytes.NewBufferString(content)
		originalMetadata := storage.Metadata{
			Filename:      "original.txt",
			ContentType:   "text/plain",
			ContentLength: "13",
			PasswordHash:  "hash123",
			ExpiresAt:     "2025-12-31T23:59:59Z",
		}

		err := s.Put(ctx, filename, reader, originalMetadata)
		if err != nil {
			t.Fatalf("Failed to put file: %v", err)
		}

		// Get only metadata
		metadata, err := s.GetOnlyMetadata(ctx, filename)
		if err != nil {
			t.Fatalf("Failed to get metadata: %v", err)
		}

		// Verify metadata
		if metadata.Filename != originalMetadata.Filename {
			t.Errorf("Filename mismatch. Expected: %s, Got: %s", originalMetadata.Filename, metadata.Filename)
		}
		if metadata.ContentType != originalMetadata.ContentType {
			t.Errorf("ContentType mismatch. Expected: %s, Got: %s", originalMetadata.ContentType, metadata.ContentType)
		}
		if metadata.ContentLength != originalMetadata.ContentLength {
			t.Errorf("ContentLength mismatch. Expected: %s, Got: %s", originalMetadata.ContentLength, metadata.ContentLength)
		}
		if metadata.PasswordHash != originalMetadata.PasswordHash {
			t.Errorf("PasswordHash mismatch. Expected: %s, Got: %s", originalMetadata.PasswordHash, metadata.PasswordHash)
		}
		if metadata.ExpiresAt != originalMetadata.ExpiresAt {
			t.Errorf("ExpiresAt mismatch. Expected: %s, Got: %s", originalMetadata.ExpiresAt, metadata.ExpiresAt)
		}
	})
}

func TestLocalStorage_List(t *testing.T) {
	tempDir := t.TempDir()
	s, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	t.Run("lists files with metadata", func(t *testing.T) {
		ctx := context.Background()

		// Put multiple files
		files := []struct {
			filename string
			content  string
			metadata storage.Metadata
		}{
			{
				filename: "file1.txt",
				content:  "Content 1",
				metadata: storage.Metadata{
					Filename:      "original1.txt",
					ContentType:   "text/plain",
					ContentLength: "9",
				},
			},
			{
				filename: "file2.txt",
				content:  "Content 2",
				metadata: storage.Metadata{
					Filename:      "original2.txt",
					ContentType:   "text/plain",
					ContentLength: "9",
				},
			},
		}

		for _, f := range files {
			reader := bytes.NewBufferString(f.content)
			err := s.Put(ctx, f.filename, reader, f.metadata)
			if err != nil {
				t.Fatalf("Failed to put file %s: %v", f.filename, err)
			}
		}

		// List files
		filenames, metadata, err := s.List(ctx)
		if err != nil {
			t.Fatalf("Failed to list files: %v", err)
		}

		// Verify results
		if len(filenames) != 2 {
			t.Errorf("Expected 2 files, got %d", len(filenames))
		}
		if len(metadata) != 2 {
			t.Errorf("Expected 2 metadata entries, got %d", len(metadata))
		}

		// Verify files are in the list
		foundFiles := make(map[string]bool)
		for _, filename := range filenames {
			foundFiles[filename] = true
		}

		for _, f := range files {
			if !foundFiles[f.filename] {
				t.Errorf("File %s not found in list", f.filename)
			}
		}
	})

	t.Run("returns empty list for empty directory", func(t *testing.T) {
		// Create a separate temp directory for empty test
		emptyTempDir := t.TempDir()
		emptyStorage, err := storage.NewLocalStorage(emptyTempDir)
		if err != nil {
			t.Fatalf("Failed to create local storage: %v", err)
		}

		ctx := context.Background()

		filenames, metadata, err := emptyStorage.List(ctx)
		if err != nil {
			t.Fatalf("Failed to list files: %v", err)
		}

		if len(filenames) != 0 {
			t.Errorf("Expected 0 files, got %d", len(filenames))
		}
		if len(metadata) != 0 {
			t.Errorf("Expected 0 metadata entries, got %d", len(metadata))
		}
	})
}

func TestLocalStorage_Delete(t *testing.T) {
	tempDir := t.TempDir()
	s, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	t.Run("deletes existing file and metadata", func(t *testing.T) {
		ctx := context.Background()
		filename := "test.txt"
		content := "Hello, World!"

		// Put file first
		reader := bytes.NewBufferString(content)
		metadata := storage.Metadata{
			Filename:      "original.txt",
			ContentType:   "text/plain",
			ContentLength: "13",
		}

		err := s.Put(ctx, filename, reader, metadata)
		if err != nil {
			t.Fatalf("Failed to put file: %v", err)
		}

		// Verify file exists
		filePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("File was not created: %s", filePath)
		}

		// Delete file
		err = s.Delete(ctx, filename)
		if err != nil {
			t.Fatalf("Failed to delete file: %v", err)
		}

		// Verify file is deleted
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Errorf("File was not deleted: %s", filePath)
		}

		// Verify metadata file is deleted
		metadataPath := filePath + ".metadata"
		if _, err := os.Stat(metadataPath); !os.IsNotExist(err) {
			t.Errorf("Metadata file was not deleted: %s", metadataPath)
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		ctx := context.Background()
		filename := "non-existent.txt"

		err := s.Delete(ctx, filename)
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
}

func TestLocalStorage_DeleteExpired(t *testing.T) {
	tempDir := t.TempDir()
	s, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	t.Run("deletes expired files", func(t *testing.T) {
		ctx := context.Background()

		// Create expired file
		expiredFile := "expired.txt"
		expiredContent := "Expired content"
		expiredReader := bytes.NewBufferString(expiredContent)
		expiredMetadata := storage.Metadata{
			Filename:      "expired.txt",
			ContentType:   "text/plain",
			ContentLength: "15",
			ExpiresAt:     "2020-01-01T00:00:00Z", // Already expired
		}

		err := s.Put(ctx, expiredFile, expiredReader, expiredMetadata)
		if err != nil {
			t.Fatalf("Failed to put expired file: %v", err)
		}

		// Create non-expired file
		validFile := "valid.txt"
		validContent := "Valid content"
		validReader := bytes.NewBufferString(validContent)
		validMetadata := storage.Metadata{
			Filename:      "valid.txt",
			ContentType:   "text/plain",
			ContentLength: "13",
			ExpiresAt:     "2099-01-01T00:00:00Z", // Future date
		}

		err = s.Put(ctx, validFile, validReader, validMetadata)
		if err != nil {
			t.Fatalf("Failed to put valid file: %v", err)
		}

		// Create file without expiration
		noExpiryFile := "no-expiry.txt"
		noExpiryContent := "No expiry content"
		noExpiryReader := bytes.NewBufferString(noExpiryContent)
		noExpiryMetadata := storage.Metadata{
			Filename:      "no-expiry.txt",
			ContentType:   "text/plain",
			ContentLength: "17",
			ExpiresAt:     "", // No expiration
		}

		err = s.Put(ctx, noExpiryFile, noExpiryReader, noExpiryMetadata)
		if err != nil {
			t.Fatalf("Failed to put no-expiry file: %v", err)
		}

		// Delete expired files
		deletedCount, err := s.DeleteExpired(ctx)
		if err != nil {
			t.Fatalf("Failed to delete expired files: %v", err)
		}

		// Verify only expired file was deleted
		if deletedCount != 1 {
			t.Errorf("Expected 1 deleted file, got %d", deletedCount)
		}

		// Verify expired file is gone
		_, err = s.Get(ctx, expiredFile)
		if !s.FileNotExists(err) {
			t.Errorf("Expected expired file to be deleted")
		}

		// Verify valid file still exists
		validFileReader, validErr := s.Get(ctx, validFile)
		if validErr != nil {
			t.Errorf("Expected valid file to still exist: %v", validErr)
		} else {
			validFileReader.Close()
		}

		// Verify no-expiry file still exists
		noExpiryFileReader, noExpiryErr := s.Get(ctx, noExpiryFile)
		if noExpiryErr != nil {
			t.Errorf("Expected no-expiry file to still exist: %v", noExpiryErr)
		} else {
			noExpiryFileReader.Close()
		}
	})
}

func TestLocalStorage_FileNotExists(t *testing.T) {
	tempDir := t.TempDir()
	s, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	t.Run("returns true for file not exists error", func(t *testing.T) {
		ctx := context.Background()
		filename := "non-existent.txt"

		_, err := s.Get(ctx, filename)
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}

		if !s.FileNotExists(err) {
			t.Errorf("Expected FileNotExists to return true for error: %v", err)
		}
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		if s.FileNotExists(nil) {
			t.Error("Expected FileNotExists to return false for nil error")
		}
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		otherError := os.ErrPermission
		if s.FileNotExists(otherError) {
			t.Errorf("Expected FileNotExists to return false for permission error")
		}
	})
}
