# Security Notes

The PastureStack `hosts-file-updater` image intentionally runs as root because
the legacy sidekick contract rewrites the container's
Docker-managed `/etc/hosts` file. Running as a non-root user breaks the legacy
behavior because `/etc/hosts` is root-owned at runtime.

Compensating controls in this maintenance fork:

- The runtime image is reduced to Ubuntu 26.04 plus a static Go binary.
- No shell scripts, vendored Go libraries, Dapper binary, or Docker CLI are
  present in the runtime image.
- Legacy metadata hostnames and agent IPs are validated before writing
  `/etc/hosts`.
- The Trivy `DS-0002` root-user finding is ignored only for this compatibility
  reason. Do not copy this exception to services that do not need to write
  root-owned runtime files.
