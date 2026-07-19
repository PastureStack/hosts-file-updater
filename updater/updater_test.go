package updater

import (
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/PastureStack/hosts-file-updater/internal/metadata"
)

var upd *Updater
var client *fakeMetadataClient
var testHostsFile string

func init() {
	client = &fakeMetadataClient{
		hosts: []metadata.Host{
			{
				Hostname: "Host1",
				AgentIP:  "10.0.0.1",
			},
		},
		lock: &sync.Mutex{},
	}

	upd = &Updater{
		knownHosts:     make(map[string]string),
		MetadataClient: client,
	}
}

type fakeMetadataClient struct {
	hosts []metadata.Host
	lock  *sync.Mutex
}

func (f *fakeMetadataClient) GetHosts() ([]metadata.Host, error) {
	return f.hosts, nil
}

func TestMain(m *testing.M) {
	tmpFile, err := os.CreateTemp("", "hosts")
	if err != nil {
		panic(err)
	}
	upd.origData = "127.0.0.1    localhost localhost-ip4"
	testHostsFile = tmpFile.Name()
	upd.HostsFile = testHostsFile
	_ = tmpFile.Close()

	code := m.Run()
	_ = os.Remove(testHostsFile)
	os.Exit(code)
}

func TestDetectsHostIpChange(t *testing.T) {
	client.lock.Lock()
	client.hosts = []metadata.Host{
		{
			Hostname: "Host1",
			AgentIP:  "10.0.0.1",
		},
	}
	upd.Run("")
	hostsMap, err := parseHostsOrigFile(testHostsFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 3 {
		t.Fatalf("Expected 3 entires, found %d", len(hostsMap))
	}
	v, ok := hostsMap["Host1"]
	if !ok {
		t.Fatalf("Entry for Host1 not found after running updater service with Host1 data")
	}
	if v != "10.0.0.1" {
		t.Fatalf("Entry for Host1 not found to be 10.0.0.1 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["localhost"]
	if !ok {
		t.Fatalf("Entry for localhost not found after running updater service with localhost data")
	}
	if v != "127.0.0.1" {
		t.Fatalf("Entry for localhost not found to be as set, after running updater service with localhost data")
	}
	client.lock.Lock()
	client.hosts = []metadata.Host{
		{
			Hostname: "Host1",
			AgentIP:  "10.0.0.2",
		},
	}
	upd.Run("")
	hostsMap, err = parseHostsOrigFile(testHostsFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 3 {
		t.Fatalf("Expected 3 entires, found %d", len(hostsMap))
	}
	v, ok = hostsMap["Host1"]
	if !ok {
		t.Fatalf("Entry for Host1 not found after running updater service with Host1 data")
	}
	if v != "10.0.0.2" {
		t.Fatalf("Entry for Host1 not found to be 10.0.0.2 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["localhost"]
	if !ok {
		t.Fatalf("Entry for localhost not found after running updater service with localhost data")
	}
	if v != "127.0.0.1" {
		t.Fatalf("Entry for localhost not found to be as set, after running updater service with localhost data")
	}
}

func TestDetectsHostAddition(t *testing.T) {
	client.lock.Lock()
	client.hosts = []metadata.Host{
		{
			Hostname: "Host1",
			AgentIP:  "10.0.0.1",
		},
	}
	upd.Run("")
	hostsMap, err := parseHostsOrigFile(testHostsFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 3 {
		t.Fatalf("Expected 3 entires, found %d", len(hostsMap))
	}
	v, ok := hostsMap["Host1"]
	if !ok {
		t.Fatalf("Entry for Host1 not found after running updater service with Host1 data")
	}
	if v != "10.0.0.1" {
		t.Fatalf("Entry for Host1 not found to be 10.0.0.1 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["localhost"]
	if !ok {
		t.Fatalf("Entry for localhost not found after running updater service with localhost data")
	}
	if v != "127.0.0.1" {
		t.Fatalf("Entry for localhost not found to be as set, after running updater service with localhost data")
	}
	client.lock.Lock()
	client.hosts = []metadata.Host{
		{
			Hostname: "Host1",
			AgentIP:  "10.0.0.1",
		},
		{
			Hostname: "Host2",
			AgentIP:  "10.0.0.2",
		},
	}
	upd.Run("")
	hostsMap, err = parseHostsOrigFile(testHostsFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 4 {
		t.Fatalf("Expected 4 entires, found %d", len(hostsMap))
	}
	v, ok = hostsMap["Host1"]
	if !ok {
		t.Fatalf("Entry for Host1 not found after running updater service with Host1 data")
	}
	if v != "10.0.0.1" {
		t.Fatalf("Entry for Host1 not found to be 10.0.0.1 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["Host2"]
	if !ok {
		t.Fatalf("Entry for Host2 not found after running updater service with Host2 data")
	}
	if v != "10.0.0.2" {
		t.Fatalf("Entry for Host2 not found to be 10.0.0.2 as set, after running updater service with Host2 data")
	}
	v, ok = hostsMap["localhost"]
	if !ok {
		t.Fatalf("Entry for localhost not found after running updater service with localhost data")
	}
	if v != "127.0.0.1" {
		t.Fatalf("Entry for localhost not found to be as set, after running updater service with localhost data")
	}
}

func TestDetectsHostDeletion(t *testing.T) {
	client.lock.Lock()
	client.hosts = []metadata.Host{
		{
			Hostname: "Host1",
			AgentIP:  "10.0.0.1",
		},
		{
			Hostname: "Host2",
			AgentIP:  "10.0.0.2",
		},
	}
	upd.Run("")
	hostsMap, err := parseHostsOrigFile(testHostsFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 4 {
		t.Fatalf("Expected 4 entires, found %d", len(hostsMap))
	}
	v, ok := hostsMap["Host1"]
	if !ok {
		t.Fatalf("Entry for Host1 not found after running updater service with Host1 data")
	}
	if v != "10.0.0.1" {
		t.Fatalf("Entry for Host1 not found to be 10.0.0.1 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["Host2"]
	if !ok {
		t.Fatalf("Entry for Host2 not found after running updater service with Host2 data")
	}
	if v != "10.0.0.2" {
		t.Fatalf("Entry for Host2 not found to be 10.0.0.2 as set, after running updater service with Host2 data")
	}
	v, ok = hostsMap["localhost"]
	if !ok {
		t.Fatalf("Entry for localhost not found after running updater service with localhost data")
	}
	if v != "127.0.0.1" {
		t.Fatalf("Entry for localhost not found to be as set, after running updater service with localhost data")
	}
	client.lock.Lock()
	client.hosts = []metadata.Host{
		{
			Hostname: "Host1",
			AgentIP:  "10.0.0.1",
		},
	}
	upd.Run("")
	hostsMap, err = parseHostsOrigFile(testHostsFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 3 {
		t.Fatalf("Expected 3 entires, found %d", len(hostsMap))
	}
	v, ok = hostsMap["Host1"]
	if !ok {
		t.Fatalf("Entry for Host1 not found after running updater service with Host1 data")
	}
	if v != "10.0.0.1" {
		t.Fatalf("Entry for Host1 not found to be 10.0.0.1 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["localhost"]
	if !ok {
		t.Fatalf("Entry for localhost not found after running updater service with localhost data")
	}
	if v != "127.0.0.1" {
		t.Fatalf("Entry for localhost not found to be as set, after running updater service with localhost data")
	}
}

func parseHostsOrigFile(file string) (map[string]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	hostsMap := map[string]string{}
	lines := string(data)
	for _, line := range strings.Split(lines, "\n") {
		elements := strings.Split(line, "    ")
		if len(elements) < 2 {
			continue
		}
		names := strings.Split(strings.Trim(elements[1], " "), " ")
		for _, name := range names {
			hostsMap[name] = elements[0]
		}
	}
	return hostsMap, nil
}

func TestRejectsUnsafeMetadataValues(t *testing.T) {
	client.lock.Lock()
	defer client.lock.Unlock()

	client.hosts = []metadata.Host{
		{
			Hostname: "good-host",
			AgentIP:  "10.0.0.10",
		},
		{
			Hostname: "bad\n10.9.9.9 owned",
			AgentIP:  "10.0.0.11",
		},
		{
			Hostname: "bad-ip",
			AgentIP:  "not-an-ip",
		},
	}
	upd.knownHosts = map[string]string{}
	upd.Run("")

	hostsMap, err := parseHostsOrigFile(testHostsFile)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if hostsMap["good-host"] != "10.0.0.10" {
		t.Fatalf("expected safe host to be written, found %q", hostsMap["good-host"])
	}
	if _, ok := hostsMap["owned"]; ok {
		t.Fatalf("unsafe newline-injected hostname was written")
	}
	if _, ok := hostsMap["bad-ip"]; ok {
		t.Fatalf("host with invalid agent IP was written")
	}
}
