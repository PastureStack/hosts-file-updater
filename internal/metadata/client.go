package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const maxMetadataResponseBytes = 16 << 20

// Host is the subset of the legacy metadata API used by hosts-file-updater.
type Host struct {
	Name     string            `json:"name"`
	AgentIP  string            `json:"agent_ip"`
	HostID   int               `json:"host_id"`
	Labels   map[string]string `json:"labels"`
	UUID     string            `json:"uuid"`
	Hostname string            `json:"hostname"`
}

// Client implements the legacy metadata calls required by this service.
type Client struct {
	baseURL string
	client  *http.Client
}

// NewClientAndWait creates a metadata client and waits for metadata to answer.
func NewClientAndWait(baseURL string, timeout time.Duration) (*Client, error) {
	client := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
	}

	if err := testConnection(client); err != nil {
		return nil, err
	}

	return client, nil
}

// OnChange polls the metadata version and invokes do when metadata changes.
func (m *Client) OnChange(intervalSeconds int, do func(string)) {
	interval := time.Duration(intervalSeconds) * time.Second
	version := "init"

	for {
		newVersion, err := m.GetVersion()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error reading metadata version: %v\n", err)
			time.Sleep(interval)
			continue
		}

		if version == newVersion {
			time.Sleep(interval)
			continue
		}

		version = newVersion
		do(newVersion)
	}
}

// GetVersion returns the current metadata version string.
func (m *Client) GetVersion() (string, error) {
	body, err := m.sendRequest(context.Background(), "/version")
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// GetHosts returns all hosts from the metadata service.
func (m *Client) GetHosts() ([]Host, error) {
	body, err := m.sendRequest(context.Background(), "/hosts")
	if err != nil {
		return nil, err
	}

	var hosts []Host
	if err := json.Unmarshal(body, &hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

func (m *Client) sendRequest(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("metadata status %d accessing %s", resp.StatusCode, path)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxMetadataResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxMetadataResponseBytes {
		return nil, fmt.Errorf("metadata response exceeded %d bytes for %s", maxMetadataResponseBytes, path)
	}
	return body, nil
}

func testConnection(mdClient *Client) error {
	var err error
	for delay := time.Second; delay < 20*time.Second; delay *= 2 {
		if _, err = mdClient.GetVersion(); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return err
}
