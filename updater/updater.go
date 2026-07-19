package updater

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/PastureStack/hosts-file-updater/internal/metadata"
)

// MetadataClient - This abstraction allows this to be mocked easily in tests
type MetadataClient interface {
	GetHosts() ([]metadata.Host, error)
}

type Updater struct {
	MetadataClient MetadataClient
	HostsFile      string
	knownHosts     map[string]string
	origData       string
}

func (u *Updater) Run(string) {
	if u.knownHosts == nil {
		u.knownHosts = make(map[string]string)
	}
	if u.origData == "" {
		u.origData = `127.0.0.1    localhost
::1    localhost ip6-localhost ip6-loopback
fe00::0    ip6-localnet
ff00::0    ip6-mcastprefix
ff02::1    ip6-allnodes
ff02::2    ip6-allrouters
`

		hostname, err := os.Hostname()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error getting hostname of host: %v\n", err)
			return
		}
		ips, err := net.LookupIP(hostname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error getting IP addresses of host %s, err: %v\n", hostname, err)
			return
		}
		if len(ips) == 0 {
			fmt.Fprintf(os.Stderr, "ERROR: Error getting IP address of host %s, err: No IPs found\n", hostname)
			return
		}
		u.origData = u.origData + ips[0].String() + "    " + hostname
	}
	err := u.Update(u.knownHosts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error updating %s: [%v]\n", u.hostsFile(), err)
	}
}

func (u *Updater) Update(knownHosts map[string]string) error {
	hosts, err := u.MetadataClient.GetHosts()
	if err != nil {
		return err
	}

	changed := false

	hostsMap := map[string]string{}

	for _, host := range hosts {
		if !validHostname(host.Hostname) {
			fmt.Fprintf(os.Stderr, "WARN: Ignoring metadata host with unsafe hostname %q\n", host.Hostname)
			continue
		}
		if net.ParseIP(host.AgentIP) == nil {
			fmt.Fprintf(os.Stderr, "WARN: Ignoring metadata host %q with invalid agent IP %q\n", host.Hostname, host.AgentIP)
			continue
		}
		if _, ok := hostsMap[host.Hostname]; ok {
			// Do not add subsequent hosts with the
			// duplicate hostnames
			continue
		}
		if ip, ok := knownHosts[host.Hostname]; !ok || ip != host.AgentIP {
			// If the current host is not a part of the
			// previous set of known hosts, then a new host
			// was added
			changed = true
			fmt.Fprintf(os.Stdout, "INFO: Adding Host %s %s\n", host.Hostname, host.AgentIP)
		}
		hostsMap[host.Hostname] = host.AgentIP
	}

	for knownHost := range knownHosts {
		// a host was deleted
		if _, ok := hostsMap[knownHost]; !ok {
			fmt.Fprintf(os.Stdout, "INFO: Deleting host %s\n", knownHost)
			changed = true
		}
	}

	if len(knownHosts) != len(hostsMap) {
		changed = true
	}

	if !changed {
		return err
	}

	// Synchronize knownHosts to the current view from the metadata service.
	for k := range knownHosts {
		delete(knownHosts, k)
	}

	for k, v := range hostsMap {
		knownHosts[k] = v
	}

	toWrite := u.origData + "\n"

	hostnames := make([]string, 0, len(hostsMap))
	for hostname := range hostsMap {
		hostnames = append(hostnames, hostname)
	}
	sort.Strings(hostnames)

	for _, hostname := range hostnames {
		toWrite = toWrite + fmt.Sprintf("%s    %s\n", hostsMap[hostname], hostname)
	}

	return os.WriteFile(u.hostsFile(), []byte(toWrite), 0644)
}

func (u *Updater) hostsFile() string {
	if u.HostsFile != "" {
		return u.HostsFile
	}
	return "/etc/hosts"
}

func validHostname(hostname string) bool {
	if hostname == "" || len(hostname) > 255 {
		return false
	}
	for _, r := range hostname {
		if r <= 32 || r == 127 || strings.ContainsRune("#/\\", r) {
			return false
		}
	}
	return true
}
