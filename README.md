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

LocalOps is in product definition. There is no usable release yet. The first
workflow, supported platform, interface, implementation stack, data lifecycle,
and license have not been selected. Implementation will begin after those
choices and their trust boundaries have concrete acceptance checks.

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
