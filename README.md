# LocalOps

LocalOps is a local-first developer operations workspace for keeping multiple
projects understandable from one place.

The project is aimed at a practical set of questions:

- What projects are on this machine?
- What is running, and what owns each process or port?
- Which explicit commands operate a project?
- Why is a project unhealthy?
- Which environment keys, executable paths, repository state, dependencies, or
  delivery signals have drifted?

The initial direction is a single-developer workflow on one workstation. Core
work should remain useful offline, inspect before it controls anything, and
require explicit approval before running project commands. Secrets, environment
values, credentials, and private log content are not inventory.

## Status

LocalOps is at its first read-only foundation. It is not yet a usable release
or running dashboard.

The current Go package inspects one absolute project path selected by the user.
It reports the cleaned selected path, its direct `.git` marker path, and whether
that marker is a directory or regular file. It does not search for projects,
read Git metadata contents, execute Git or repository commands, or retain
state.

The initial product contract is Windows-first, single-user, offline-capable,
and limited to a loopback-only web dashboard. Scanning, persistence, command
execution, outbound network access, secret or environment-value collection,
log capture, and process inspection or control are outside the current slice.

## Working principles

- Local value should not require an account or hosted service.
- Discovery must be bounded to locations the user selects.
- Repository content is untrusted until the user approves an operation.
- Every observation and action should show its source.
- Unknown, stale, unavailable, unsupported, and unhealthy are different states.
- Platform differences should remain visible instead of being hidden behind
  false portability.
- New integrations should solve a complete workflow before they become public
  extension points.

The scope can grow toward project inventory, service lifecycle, operational
diagnostics, delivery integrations, and additional machines, but each stage
must earn that responsibility by solving a real workflow first.

## Development

Run the current test suite with:

```powershell
go test ./...
```

LocalOps is licensed under the [MIT License](LICENSE).
