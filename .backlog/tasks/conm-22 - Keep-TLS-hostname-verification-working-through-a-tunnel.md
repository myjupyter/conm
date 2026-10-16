---
id: CONM-22
title: Keep TLS hostname verification working through a tunnel
status: To Do
assignee: []
created_date: '2026-10-09 16:33'
labels: []
dependencies:
  - CONM-14
references:
  - CONM-4
priority: medium
type: bug
ordinal: 18000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A connection reached through a transport is dialled at the forwarded local address (127.0.0.1:<port>), so drivers and CLI clients verify the server certificate against 127.0.0.1 instead of the configured host. Modes that check the hostname — postgres verify-full, mysql VERIFY_IDENTITY, mssql/clickhouse/mongodb/redis equivalents — fail through a tunnel even though they work direct.

The configured host has to reach the TLS layer as the server name (SNI / verification name) while the TCP dial goes to the forwarded address.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A tunnelled connection with hostname verification enabled verifies against the configured host, not the forwarded address
- [ ] #2 Ping and Run both honour it, for every database type that supports TLS
- [ ] #3 A type whose driver or client cannot separate dial address from verification name is documented as a known limitation
<!-- AC:END -->
