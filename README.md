# PastureStack Hosts File Updater

`hosts-file-updater` populates a managed container's `/etc/hosts` file from a
legacy metadata service. This repository is maintained by the PastureStack
community as a compatibility project.

## Independent Community Project

PastureStack is an independent community effort to preserve, audit, and modernize the Rancher 1.6 ecosystem. It is not affiliated with or endorsed by Rancher Labs or SUSE.

**Upstream:** [`rancher/etc-host-updater`](https://github.com/rancher/etc-host-updater). This GitHub fork retains the upstream Git history, authorship, dates, and license notices unchanged; PastureStack maintenance is consolidated into one commit after the preserved upstream boundary.

Third-party names and marks remain the property of their respective owners.
See [ORIGIN.md](ORIGIN.md), [COMPATIBILITY.md](COMPATIBILITY.md), and the
unchanged [NOTICE](NOTICE) for provenance and compatibility details.

## Names and Compatibility

The preferred executable is `hosts-file-updater`. Builds also provide an
`etc-host-updater` executable alias because existing catalog templates invoke
that exact legacy command.

The default metadata base URL remains:

```text
http://metadata/2015-12-19
```

Use `--metadata-url` to point a development or isolated deployment at a mock
service without changing the legacy default:

```sh
./bin/hosts-file-updater \
  --metadata-url http://127.0.0.1:18080/2015-12-19
```

The service polls `/version`, reads `/hosts`, validates hostnames and agent IP
addresses, and writes the resulting host entries to `/etc/hosts`.

## Build

Build the preferred executable and compatibility alias on Linux:

```sh
./scripts/build
```

The outputs are:

```text
bin/hosts-file-updater
bin/etc-host-updater -> hosts-file-updater
```

The maintained container image name is:

```text
ghcr.io/pasturestack/hosts-file-updater
```

This repository currently has no GitHub Release or Catalog-pinned release. The
unversioned name above identifies the maintained package namespace only; it is
not a floating production-version promise.

## Safe Quick Test

The test suite uses an in-process mock metadata server and a temporary hosts
file. It does not write the development machine's `/etc/hosts`:

```sh
go test ./...
```

## Production Runtime Warning

Normal service execution rewrites `/etc/hosts` and therefore intentionally
runs as root in the legacy sidekick deployment model. Do not run the service
directly on a development host. Test it in an isolated container or VM, and
read [SECURITY.md](SECURITY.md) before deployment.

## License and Original Attribution

The project remains licensed under the Apache License, Version 2.0. The
existing `LICENSE` and `NOTICE` files are preserved.

PastureStack's later modifications do not remove or replace the original
authors' copyright, attribution, or project history. Exact legal attribution is
preserved in [NOTICE](NOTICE) and [ORIGIN.md](ORIGIN.md).
