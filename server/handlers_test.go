package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/exler/fileigloo/server"
	"github.com/exler/fileigloo/storage"
)

func setupTestServer(t *testing.T, maxUploadSizeMB ...int64) (*httptest.Server, *storage.LocalStorage) {
	t.Helper()

	// Default max upload size is 10MB
	maxUploadSize := int64(10)
	if len(maxUploadSizeMB) > 0 {
		maxUploadSize = maxUploadSizeMB[0]
	}

	// Create temporary directory for test storage
	tmpDir, err := os.MkdirTemp("", "fileigloo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	// Create local storage
	localStorage, err := storage.NewLocalStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	// Create server instance
	srv := server.New(
		server.UseStorage(localStorage),
		server.MaxUploadSize(maxUploadSize),
		server.MaxRequests(100),
		server.Port(0), // Use random port
	)

	// Create test server
	testServer := httptest.NewServer(srv.GetRouter())
	t.Cleanup(testServer.Close)

	return testServer, localStorage
}

func TestFileUploadHandler(t *testing.T) {
	t.Run("upload file with JSON response", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		// Create a test file
		fileContent := "Hello, World!"

		// Create multipart form
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		// Add file field
		fileField, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		_, err = fileField.Write([]byte(fileContent))
		if err != nil {
			t.Fatalf("Failed to write file content: %v", err)
		}

		// Add optional fields
		writer.WriteField("password", "test123")
		writer.WriteField("expiration", "24")

		writer.Close()

		// Make request to server
		req, err := http.NewRequest("POST", ts.URL+"/", &buf)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Verify JSON response
		var uploadResp server.FileUploadResponse
		if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		if uploadResp.FileId == "" {
			t.Error("Expected non-empty FileId")
		}

		if uploadResp.FileUrl == "" {
			t.Error("Expected non-empty FileUrl")
		}

		if !strings.Contains(uploadResp.FileUrl, uploadResp.FileId) {
			t.Errorf("Expected FileUrl to contain FileId, got %s", uploadResp.FileUrl)
		}
	})

	t.Run("upload text with JSON response", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		// Create form data
		formData := url.Values{}
		formData.Set("text", "Hello, World!")
		formData.Set("password", "test123")
		formData.Set("expiration", "12")

		// Make request to server
		req, err := http.NewRequest("POST", ts.URL+"/", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Verify JSON response
		var uploadResp server.FileUploadResponse
		if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		if uploadResp.FileId == "" {
			t.Error("Expected non-empty FileId")
		}

		if uploadResp.FileUrl == "" {
			t.Error("Expected non-empty FileUrl")
		}
	})

	t.Run("reject both file and text", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		// Create multipart form with both file and text
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		// Add file field
		fileField, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		fileField.Write([]byte("file content"))

		// Add text field
		writer.WriteField("text", "text content")

		writer.Close()

		// Make request to server
		req, err := http.NewRequest("POST", ts.URL+"/", &buf)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("reject missing file and text", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		// Create empty form
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		writer.Close()

		// Make request to server
		req, err := http.NewRequest("POST", ts.URL+"/", &buf)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("file too large", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		// Create a file larger than max upload size (10MB in setup)
		largeContent := make([]byte, 11*1024*1024) // 11MB

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		fileField, err := writer.CreateFormFile("file", "large.bin")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		_, err = fileField.Write(largeContent)
		if err != nil {
			t.Fatalf("Failed to write file content: %v", err)
		}

		writer.Close()

		// Make request to server
		req, err := http.NewRequest("POST", ts.URL+"/", &buf)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusRequestEntityTooLarge {
			t.Errorf("Expected status 413, got %d", resp.StatusCode)
		}
	})

	t.Run("upload 100MB file successfully", func(t *testing.T) {
		// Create server with 200MB max upload size
		ts, _ := setupTestServer(t, 200)

		// Create a 100MB file
		fileSize := 100 * 1024 * 1024 // 100MB
		largeContent := make([]byte, fileSize)
		// Fill with some pattern to ensure it's not just zeros
		for i := range largeContent {
			largeContent[i] = byte(i % 256)
		}

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		fileField, err := writer.CreateFormFile("file", "large-100mb.bin")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		_, err = fileField.Write(largeContent)
		if err != nil {
			t.Fatalf("Failed to write file content: %v", err)
		}

		writer.Close()

		// Make request to server
		req, err := http.NewRequest("POST", ts.URL+"/", &buf)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// Verify JSON response
		var uploadResp server.FileUploadResponse
		if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		if uploadResp.FileId == "" {
			t.Error("Expected non-empty FileId")
		}

		if uploadResp.FileUrl == "" {
			t.Error("Expected non-empty FileUrl")
		}

		// Verify the file was actually stored by downloading it
		downloadURL := uploadResp.FileUrl
		downloadURL = strings.Replace(downloadURL, "/view/", "/download/", 1)

		downloadReq, err := http.NewRequest("GET", downloadURL, nil)
		if err != nil {
			t.Fatalf("Failed to create download request: %v", err)
		}

		downloadResp, err := http.DefaultClient.Do(downloadReq)
		if err != nil {
			t.Fatalf("Failed to make download request: %v", err)
		}
		defer downloadResp.Body.Close()

		if downloadResp.StatusCode != http.StatusOK {
			t.Errorf("Expected download status 200, got %d", downloadResp.StatusCode)
		}

		// Verify the downloaded file size matches
		downloadedContent, err := io.ReadAll(downloadResp.Body)
		if err != nil {
			t.Fatalf("Failed to read downloaded content: %v", err)
		}

		if len(downloadedContent) != fileSize {
			t.Errorf("Expected downloaded file size %d bytes, got %d bytes", fileSize, len(downloadedContent))
		}

		// Verify content matches (check first and last KB to avoid full comparison overhead)
		if !bytes.Equal(downloadedContent[:1024], largeContent[:1024]) {
			t.Error("First 1KB of downloaded content doesn't match uploaded content")
		}

		if !bytes.Equal(downloadedContent[len(downloadedContent)-1024:], largeContent[len(largeContent)-1024:]) {
			t.Error("Last 1KB of downloaded content doesn't match uploaded content")
		}
	})
}

func TestDownloadHandler(t *testing.T) {
	t.Run("download file", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		// First, upload a file
		fileContent := "Test file content"

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		fileField, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		fileField.Write([]byte(fileContent))
		writer.Close()

		req, err := http.NewRequest("POST", ts.URL+"/", &buf)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var uploadResp server.FileUploadResponse
		if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		// Now download the file
		downloadURL := strings.Replace(uploadResp.FileUrl, "/view/", "/download/", 1)
		downloadReq, err := http.NewRequest("GET", downloadURL, nil)
		if err != nil {
			t.Fatalf("Failed to create download request: %v", err)
		}

		downloadResp, err := http.DefaultClient.Do(downloadReq)
		if err != nil {
			t.Fatalf("Failed to make download request: %v", err)
		}
		defer downloadResp.Body.Close()

		if downloadResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", downloadResp.StatusCode)
		}

		// Verify content disposition is attachment
		disposition := downloadResp.Header.Get("Content-Disposition")
		if !strings.Contains(disposition, "attachment") {
			t.Errorf("Expected Content-Disposition to contain 'attachment', got '%s'", disposition)
		}

		// Verify file content
		downloadedContent, err := io.ReadAll(downloadResp.Body)
		if err != nil {
			t.Fatalf("Failed to read downloaded content: %v", err)
		}

		if string(downloadedContent) != fileContent {
			t.Errorf("Expected content '%s', got '%s'", fileContent, string(downloadedContent))
		}
	})

	t.Run("view text paste inline", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		// Upload text which should be viewable inline
		textContent := "This is a text paste"

		formData := url.Values{}
		formData.Set("text", textContent)

		req, err := http.NewRequest("POST", ts.URL+"/", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var uploadResp server.FileUploadResponse
		if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		// Text pastes should use /view/ URL
		if !strings.Contains(uploadResp.FileUrl, "/view/") {
			t.Errorf("Expected text paste to have /view/ URL, got %s", uploadResp.FileUrl)
		}

		// Access the view URL
		viewReq, err := http.NewRequest("GET", uploadResp.FileUrl, nil)
		if err != nil {
			t.Fatalf("Failed to create view request: %v", err)
		}

		viewResp, err := http.DefaultClient.Do(viewReq)
		if err != nil {
			t.Fatalf("Failed to make view request: %v", err)
		}
		defer viewResp.Body.Close()

		if viewResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", viewResp.StatusCode)
		}

		// Verify content disposition is inline
		disposition := viewResp.Header.Get("Content-Disposition")
		if !strings.Contains(disposition, "inline") {
			t.Errorf("Expected Content-Disposition to contain 'inline', got '%s'", disposition)
		}

		// Verify content
		content, err := io.ReadAll(viewResp.Body)
		if err != nil {
			t.Fatalf("Failed to read content: %v", err)
		}

		if string(content) != textContent {
			t.Errorf("Expected content '%s', got '%s'", textContent, string(content))
		}
	})

	t.Run("view file", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		// Upload a file
		fileContent := "Test file content"

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		fileField, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		fileField.Write([]byte(fileContent))
		writer.Close()

		req, err := http.NewRequest("POST", ts.URL+"/", &buf)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var uploadResp server.FileUploadResponse
		if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		// Use the URL as returned (might be view or download depending on content type detection)
		viewReq, err := http.NewRequest("GET", uploadResp.FileUrl, nil)
		if err != nil {
			t.Fatalf("Failed to create view request: %v", err)
		}

		viewResp, err := http.DefaultClient.Do(viewReq)
		if err != nil {
			t.Fatalf("Failed to make view request: %v", err)
		}
		defer viewResp.Body.Close()

		if viewResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", viewResp.StatusCode)
		}

		// Verify we can read the content
		content, err := io.ReadAll(viewResp.Body)
		if err != nil {
			t.Fatalf("Failed to read content: %v", err)
		}

		if string(content) != fileContent {
			t.Errorf("Expected content '%s', got '%s'", fileContent, string(content))
		}
	})

	t.Run("download nonexistent file", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		req, err := http.NewRequest("GET", ts.URL+"/download/nonexistent", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("password protected file", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		// Upload a password-protected file
		fileContent := "Secret content"

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		fileField, err := writer.CreateFormFile("file", "secret.txt")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		fileField.Write([]byte(fileContent))
		writer.WriteField("password", "secret123")
		writer.Close()

		req, err := http.NewRequest("POST", ts.URL+"/", &buf)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var uploadResp server.FileUploadResponse
		if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		// Try to access without password - should show password form
		viewReq, err := http.NewRequest("GET", uploadResp.FileUrl, nil)
		if err != nil {
			t.Fatalf("Failed to create view request: %v", err)
		}

		viewResp, err := http.DefaultClient.Do(viewReq)
		if err != nil {
			t.Fatalf("Failed to make view request: %v", err)
		}
		defer viewResp.Body.Close()

		if viewResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 (password form), got %d", viewResp.StatusCode)
		}

		// Try with correct password
		formData := url.Values{}
		formData.Set("password", "secret123")

		postReq, err := http.NewRequest("POST", uploadResp.FileUrl, strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("Failed to create POST request: %v", err)
		}
		postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		postResp, err := http.DefaultClient.Do(postReq)
		if err != nil {
			t.Fatalf("Failed to make POST request: %v", err)
		}
		defer postResp.Body.Close()

		if postResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", postResp.StatusCode)
		}

		// Verify we got the file content
		downloadedContent, err := io.ReadAll(postResp.Body)
		if err != nil {
			t.Fatalf("Failed to read downloaded content: %v", err)
		}

		if string(downloadedContent) != fileContent {
			t.Errorf("Expected content '%s', got '%s'", fileContent, string(downloadedContent))
		}
	})
}
