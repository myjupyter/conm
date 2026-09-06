---
id: CONM-2
title: Reach a server that is not on the client's network — pluggable transports
status: To Do
assignee:
  - myjupyter
created_date: '2026-09-06 12:27'
updated_date: '2026-09-06 14:56'
due_date: '2026-09-06'
labels:
  - not-ready
milestone: m-0
dependencies: []
priority: high
type: feature
ordinal: 2000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
### Things to complete before start:

- view design
- plan and architecture
- what transport will be implemented

### Description

Every client in `internal/network` dials the host in the DSN directly over TCP. That
assumes the server is routable from the machine conm runs on, which is often false: the
database sits behind a bastion, inside a Kubernetes cluster, or behind a cloud-managed
proxy, and the address in the config is only meaningful on the far side of a tunnel.

Transports worth supporting:

- **SSH tunnel** — a jump host, key or agent auth, local port forwarding to the database.
- **Kubernetes port-forward** — a pod or a service in a namespace, through the current
  kubeconfig context.
- **Cloud proxies** — `cloud-sql-proxy` and friends, which are a local endpoint plus a
  managed identity.
- **SOCKS/HTTP proxy** — a plain `Dialer` swap for the drivers that accept one.

What it touches:

1. **The model.** A connection needs to say how it is reached, not only where. That is a
   transport reference on the connection (kind plus its own settings, the way the secret is
   a reference rather than a password) and — for SSH keys or a kubeconfig — credentials that
   are *not* the database password, which the current single `SecretRef()` cannot hold.
2. **The lifecycle.** A transport is a live thing: it must come up before `Ping`/`Run` and
   come down after, and it is shared by both the driver and the CLI child process. This
   belongs in `network`, beside the client, not in a screen.
3. **The address the client is handed.** Once a tunnel stands, the driver and the CLI must
   be pointed at the local forwarded endpoint while the UI keeps showing the real host, so
   `ConnectionString`, `OpError.Target` and `During` need to distinguish the address dialled
   from the address configured.
4. **The failures.** A transport can fail on its own (no route to the bastion, key rejected,
   pod gone) before the database is ever reached. Those are new failure sources that must
   still surface as an `*OpError` — most map onto existing codes; anything that genuinely
   does not is a new `ErrorCode` plus a `hintFor` arm, which is a deliberate decision.
<!-- SECTION:DESCRIPTION:END -->
