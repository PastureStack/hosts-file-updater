package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/PastureStack/hosts-file-updater/internal/metadata"
	"github.com/PastureStack/hosts-file-updater/updater"
)

func TestParseOptionsUsesNeutralMetadataURLByDefault(t *testing.T) {
	opts, err := parseOptions(nil, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	const wantMetadataURL = "http://metadata/2015-12-19"
	if opts.metadataURL != wantMetadataURL {
		t.Fatalf("metadata URL = %q, want %q", opts.metadataURL, wantMetadataURL)
	}
}

func TestParseOptionsAllowsMockMetadataURL(t *testing.T) {
	opts, err := parseOptions([]string{
		"-metadata-url", "http://127.0.0.1:18080/2015-12-19",
		"-metadata-timeout", "250ms",
		"-update-interval", "1",
	}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if opts.metadataURL != "http://127.0.0.1:18080/2015-12-19" {
		t.Fatalf("unexpected metadata URL %q", opts.metadataURL)
	}
	if opts.metadataTimeout != 250*time.Millisecond {
		t.Fatalf("metadata timeout = %s, want 250ms", opts.metadataTimeout)
	}
	if opts.updateInterval != 1 {
		t.Fatalf("update interval = %d, want 1", opts.updateInterval)
	}
}

func TestParseOptionsRejectsEmptyMetadataURL(t *testing.T) {
	_, err := parseOptions([]string{"-metadata-url", ""}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected an empty metadata URL to be rejected")
	}
}

func TestMockMetadataPOCWritesOnlyTemporaryHostsFile(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/2015-12-19/version", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "mock-version")
	})
	mux.HandleFunc("/2015-12-19/hosts", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"hostname":"worker-1","agent_ip":"10.0.0.10"}]`)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := metadata.NewClientAndWait(server.URL+"/2015-12-19", time.Second)
	if err != nil {
		t.Fatal(err)
	}

	hostsFile := filepath.Join(t.TempDir(), "hosts")
	u := updater.Updater{
		MetadataClient: client,
		HostsFile:      hostsFile,
	}
	if err := u.Update(map[string]string{}); err != nil {
		t.Fatal(err)
	}

	contents, err := os.ReadFile(hostsFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "10.0.0.10    worker-1") {
		t.Fatalf("temporary hosts file did not contain mock host: %q", contents)
	}
}
