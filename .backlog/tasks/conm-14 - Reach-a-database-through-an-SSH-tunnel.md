---
id: CONM-14
title: Reach a database through an SSH tunnel
status: Done
assignee: []
created_date: '2026-09-13 17:30'
updated_date: '2026-10-16 19:01'
labels: []
dependencies:
  - CONM-13
references:
  - CONM-2
documentation:
  - docs/network.md
  - docs/repository.md
  - docs/forms.md
  - CLAUDE.md
priority: high
type: feature
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Clients in internal/network dial the database host directly, which assumes it is routable from this machine. Usually it is not: the database is on a private network behind a bastion. Today the user stands the tunnel up by hand and saves a connection pointing at localhost, which loses the real address and breaks when the local port changes.

First concrete transport of CONM-2, and the one that proves the seam. The bastion is referenced, not copied: a connection names an SSH entry from CONM-13, the way a password names a secret record.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A database connection references a saved SSH entry as its tunnel; the jump host is not copied inline
- [x] #2 A dangling tunnel reference fails validation and the connection is shown unusable, never dialled direct
- [x] #3 The tunnel comes up before Ping/Run and down when the operation ends, including on CLI exit and on failure
- [x] #4 Driver and CLI client share the same tunnel instance
- [x] #5 The local forwarded port is picked automatically and two live tunnels do not collide
- [x] #6 Table, info screen, connection string and OpError target show the configured host, never the forwarded address
- [x] #7 Tunnel failures surface as *network.OpError distinguishable from database failures, with a hint naming the bastion
- [x] #8 Editing or removing an SSH entry other connections tunnel through is refused, not left dangling
- [x] #9 The tunnel is configured through the existing form spec, with no bespoke screen code
- [x] #10 Tests cover reference resolution, dangling references and tunnel-failure error mapping
- [ ] #11 docs/network.md and docs/repository.md cover the tunnel lifecycle, the dialled-vs-configured address split and cross-entity references
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verification 2026-10-16. go test ./... green (728). Live e2e against test/transport/docker-compose.yml (postgres behind linuxserver/openssh-server, password auth): Ping through ssh:jump OK (~7-10 ms); Run opened psql through the tunnel, `select inet_server_addr()` returned the db container's docker-network address (172.20.0.3), and no ssh process was left after psql exited. With the bastion stopped, Ping returned *network.OpError{Op: ping, Code: ETRANSPORT, Target: "ssh:jump", During: "dial tcp db:5432"} — the configured address, not the forwarded one; message names the bastion address (via 127.0.0.1:2222).

Design deviated from the AC wording: a connection references a Transport record (transport.toml: name, kind, via = SSH entry name, local_port, timeouts), and the Transport names the SSH entry. The jump host is still never copied inline. The Transport entity has its own table/form screens, following the secrets-screen pattern; the connection's transport field is spec-driven (transportTypeField/transportField).

Open gaps: AC#2 — a dangling ref is refused at Add/Edit (checkTransport), and a file-loaded dangling ref is never dialled direct (Ping -> ETRANSPORT "transport ... does not exist"), but the row is not marked unusable before a ping and the hint reads "did not come up". AC#7 — ETRANSPORT hint is generic; the bastion appears only as host:port in the error text, not by entry name. AC#11 — docs are done separately.

2026-10-16: the author accepted AC#2 and AC#7 as they are (a dangling ref is refused at the write point and never dialled direct; the transport failure text names the bastion address). AC#11 is left to the separate docs effort.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
A database connection can be reached through a saved SSH entry.

What changed (commit c4043db):
- `config.Transport` + `transport.toml`: a named transport record (`kind = ssh`, `via` = SSH entry name, optional `local_port`, `connect_timeout`, `keepalive`). Database configs carry `transport = "ssh:<name>"`; the bastion is referenced, not copied inline.
- `internal/transport`: `SSH` brings up an `ssh -L` forward on an auto-reserved local port, waits for it to be ready, and refcounts users so a ping and the CLI share one live tunnel. `Pool` shares a tunnel per (transport, target).
- `network.Transported` / `NewRouted`: opens the transport before Ping/Run and closes it after, including on CLI exit and on failure. It rewrites `OpError.Target`/`During` back to the configured address and sets `Through`. Transport failures map to `ETRANSPORT`, `EADDRINUSE` or the secret code.
- `repository.TransportRepository`, plus `TransportIndex`/`ConnectionIndex` usage indexes: editing or removing a transport, or the SSH entry it goes through, while something uses it is refused with `TransportInUseError`/`ConnectionInUseError`. A dangling ref is refused at Add/Edit.
- UI: a transport table and form (same pattern as secrets); the connection's transport field comes from the spec.

Verification: `go test ./...` passes (728 tests). Live e2e on `test/transport/docker-compose.yml`:
- Ping through the bastion succeeds.
- `psql` through the tunnel reached the db on its docker-network address, and no ssh process was left after exit.
- With the bastion stopped, Ping returns ETRANSPORT and shows the configured target.

Accepted limitations: a dangling ref loaded from the file is only reported on ping. The ETRANSPORT hint is generic; the bastion appears as host:port in the error text. Docs (AC#11) belong to the separate docs effort.
<!-- SECTION:FINAL_SUMMARY:END -->
