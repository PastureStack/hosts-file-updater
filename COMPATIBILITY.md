# Compatibility Contract

This document separates PastureStack naming from interfaces that must remain
compatible with existing legacy 1.6 deployments.

## Preserved Runtime Contracts

The following values and behavior are intentionally retained:

- Default metadata base URL:
  `http://metadata/2015-12-19`.
- Metadata endpoints: `/version` and `/hosts`.
- Metadata host fields including `hostname`, `agent_ip`, `host_id`, `labels`,
  `uuid`, and `name`.
- The root-owned `/etc/hosts` output path used by the legacy sidekick model.
- The `etc-host-updater` executable alias used by existing catalog templates.
- Existing `--update-interval`, `--metadata-timeout`, `--debug`, and
  `--version` command-line options.

The executable alias is a compatibility identifier, not PastureStack branding.
The neutral metadata hostname is the new default for PastureStack deployments.

## PastureStack Interfaces

The preferred interfaces are:

- Repository and module: `github.com/PastureStack/hosts-file-updater`.
- Executable: `hosts-file-updater`.
- Container image: `ghcr.io/pasturestack/hosts-file-updater`.

The `--metadata-url` option may override the default base URL for isolated
testing or controlled deployments. Existing installations that still use a
historical DNS alias must either provide the neutral `metadata` alias or pass
their current endpoint explicitly during migration.

## POC Safety Boundary

Unit and quick POC tests use `httptest` for metadata and `t.TempDir()` for the
hosts file. They must not write the workstation's `/etc/hosts` and must not
contact a production metadata service.

An isolated runtime POC must verify both executable names, the default URL,
the override URL, `/version` polling, `/hosts` parsing, hostname and IP
validation, and `/etc/hosts` output inside a disposable Linux VM or container.

## Non-Goals

This command-line system component has no graphical UI or runtime localization
framework. Runtime i18n and translated log output are outside this migration
POC; public repository documentation is maintained in English.
