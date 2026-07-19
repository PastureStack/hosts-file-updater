package metadata

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientUsesOverrideURLForLegacyMetadataAPI(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/2015-12-19/version", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept header = %q, want application/json", got)
		}
		_, _ = fmt.Fprint(w, "mock-version")
	})
	mux.HandleFunc("/2015-12-19/hosts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"hostname":"worker-1","agent_ip":"10.0.0.10"}]`)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewClientAndWait(server.URL+"/2015-12-19", time.Second)
	if err != nil {
		t.Fatal(err)
	}

	version, err := client.GetVersion()
	if err != nil {
		t.Fatal(err)
	}
	if version != "mock-version" {
		t.Fatalf("version = %q, want mock-version", version)
	}

	hosts, err := client.GetHosts()
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 1 || hosts[0].Hostname != "worker-1" || hosts[0].AgentIP != "10.0.0.10" {
		t.Fatalf("unexpected hosts response: %#v", hosts)
	}
}
