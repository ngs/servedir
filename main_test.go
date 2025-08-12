package main

import (
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	apachelog "github.com/lestrrat-go/apache-logformat"
)

func TestFileServer(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"
	
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Create a test server
	handler := apachelog.CombinedLog.Wrap(http.FileServer(http.Dir(tmpDir)), io.Discard)
	server := httptest.NewServer(handler)
	defer server.Close()
	
	// Test serving the file
	resp, err := http.Get(server.URL + "/test.txt")
	if err != nil {
		t.Fatalf("Failed to GET test file: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	
	if string(body) != testContent {
		t.Errorf("Expected body %q, got %q", testContent, string(body))
	}
}

func TestFlags(t *testing.T) {
	// Save original command-line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	
	// Reset flag package
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	
	tests := []struct {
		name     string
		args     []string
		wantPort int
		wantErr  bool
	}{
		{
			name:     "default port (any available)",
			args:     []string{"servedir"},
			wantPort: 0,
			wantErr:  false,
		},
		{
			name:     "custom port",
			args:     []string{"servedir", "-port", "9000"},
			wantPort: 9000,
			wantErr:  false,
		},
		{
			name:     "with directory",
			args:     []string{"servedir", "-port", "9000", "/tmp"},
			wantPort: 9000,
			wantErr:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			
			var port int
			flag.IntVar(&port, "port", 0, "HTTP Port to Listen (0 for any available port)")
			
			os.Args = tt.args
			err := flag.CommandLine.Parse(os.Args[1:])
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if port != tt.wantPort {
				t.Errorf("port = %d, want %d", port, tt.wantPort)
			}
		})
	}
}

func TestDirectoryServing(t *testing.T) {
	// Create a temporary directory with subdirectories
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// Create files in different locations
	files := map[string]string{
		"index.html":        "<h1>Welcome</h1>",
		"subdir/test.txt":   "Subdirectory content",
		"subdir/index.html": "<h1>Subdirectory</h1>",
	}
	
	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}
	
	// Create a test server
	handler := http.FileServer(http.Dir(tmpDir))
	server := httptest.NewServer(handler)
	defer server.Close()
	
	// Test serving different paths
	tests := []struct {
		path     string
		wantCode int
		wantBody string
	}{
		{"/", http.StatusOK, "<h1>Welcome</h1>"},
		{"/index.html", http.StatusOK, "<h1>Welcome</h1>"},
		{"/subdir/", http.StatusOK, "<h1>Subdirectory</h1>"},
		{"/subdir/test.txt", http.StatusOK, "Subdirectory content"},
		{"/nonexistent", http.StatusNotFound, ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := http.Get(server.URL + tt.path)
			if err != nil {
				t.Fatalf("Failed to GET %s: %v", tt.path, err)
			}
			defer resp.Body.Close()
			
			if resp.StatusCode != tt.wantCode {
				t.Errorf("Path %s: expected status %d, got %d", tt.path, tt.wantCode, resp.StatusCode)
			}
			
			if tt.wantBody != "" {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				
				if string(body) != tt.wantBody {
					t.Errorf("Path %s: expected body %q, got %q", tt.path, tt.wantBody, string(body))
				}
			}
		})
	}
}

func TestConcurrentRequests(t *testing.T) {
	// Create a temporary directory with a test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "concurrent.txt")
	if err := os.WriteFile(testFile, []byte("concurrent test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Create a test server
	handler := http.FileServer(http.Dir(tmpDir))
	server := httptest.NewServer(handler)
	defer server.Close()
	
	// Make concurrent requests
	const numRequests = 50
	done := make(chan bool, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(server.URL + "/concurrent.txt")
			if err != nil {
				t.Errorf("Request failed: %v", err)
			} else {
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			}
			done <- true
		}()
	}
	
	// Wait for all requests to complete with timeout
	timeout := time.After(5 * time.Second)
	for i := 0; i < numRequests; i++ {
		select {
		case <-done:
			// Request completed
		case <-timeout:
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}
}