---
id: CONM-1
title: >-
  Show the topology of a database and the relations between connections of the
  same type
status: To Do
assignee:
  - myjupyter
created_date: '2026-09-06 12:23'
updated_date: '2026-09-06 14:56'
due_date: '2026-09-06'
labels:
  - not-ready
milestone: m-0
dependencies: []
priority: low
type: feature
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`config.Connection` describes exactly one endpoint — one `Host()`, one `Port()` — and the
connections table draws each entry as an independent row. Real deployments are not flat: a
replica set, a sharded cluster or a primary with its read replicas is several addresses
belonging to one logical database, and today conm can neither store nor show that.

Per type:

- MongoDB — a replica set is a seed list (`mongodb://h1,h2,h3/db?replicaSet=rs0`) or a DNS
  SRV record (`mongodb+srv://…`, no port at all); only the single-host form is supported.
- ClickHouse — a cluster is a list of shard and replica hosts; only the single-host form is
  supported.
- Postgres, MySQL, MSSQL — a primary and its replicas are saved as unrelated rows, with
  nothing saying which is which or that they belong together.

Two independent halves:

1. **The model.** A connection carrying more than one address, and a relation between
   connections of the same type (member of a set, replica of a primary, node of a cluster).
   This changes the shared model, not one type: the cheapest shape is an optional interface
   a client asserts (`Hosts() []string`), the honest one is a wider `config.Connection`.
   Needs a decision before any code — see the deviation table in `docs/config.md`.
2. **The view.** Grouped or nested rows under the logical database, the role of each node
   (primary/secondary/shard), and a ping that says which member answered. `ui/view` owns the
   rows and the cursor, so the grouping belongs there and not in a screen.
<!-- SECTION:DESCRIPTION:END -->
