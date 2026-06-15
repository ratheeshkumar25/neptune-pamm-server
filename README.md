# Neptune-Pamm-Server

- **Author:** Ratheesh G Kumar
- **Role:** Backend Engineer (Golang)

## Overview

**Neptune PAMM** is a percentage-allocation money-management platform for **MetaTrader 5**
brokers, powered by and licensed under 9dot technology. Money managers trade a single master
account; Neptune's allocation engine — which runs **independently of the trading server** to keep
MT5 load low — splits every deal across connected investors by balance or equity, distributes
profit, and settles fees automatically. Investor funds, manager economics, and partner commissions
stay cleanly separated end to end.

> **Platform scope:** MT5 only. The MT5 Manager API is a Windows-only C++ SDK, so platform
> connectivity lives in an external connector that webhooks events in and receives balance-op
> write-backs over websocket — never linked into this Go service.

## Highlights

- **Native MT5 integration** — real-time trade, balance, and equity sync via the MT5 Manager API;
  deal-based accounting with netting and hedging support.
- **Flexible fees** — performance (with High-Water Mark), management, annual management, connection,
  and per-lot / per-deal charges.
- **Multi-tier IB commissions** — investor-funded introducing-broker fees, attributed on each
  investor's allocated volume and split across IB tiers, fully decoupled from money-manager payouts.
- **Risk controls** — per-investor max-loss / max-profit / investments-loss limits with automatic
  disconnect.
- **Built for brokers** — full REST API, CRM / back-office integration, AML/KYC, and configurable
  statements.

## Architecture

The service follows **Clean Architecture**: dependencies point **inward**, and inner layers (domain,
use cases) never import outer layers (DB, transport, frameworks). Outer layers depend on inner ones
only through interfaces, wired together at a single composition root (`internal/di`).

```
        ┌─────────────────────────────────────────────┐
        │  cmd / pkg / internal/db   (frameworks)      │  ← drivers, I/O
        │   ┌─────────────────────────────────────┐    │
        │   │ handler · repo · proto  (adapters)   │    │  ← delivery & persistence
        │   │   ┌─────────────────────────────┐    │    │
        │   │   │   services   (use cases)     │    │    │  ← application logic
        │   │   │    ┌───────────────────┐     │    │    │
        │   │   │    │  model (entities) │     │    │    │  ← pure domain
        │   │   │    └───────────────────┘     │    │    │
        │   │   └─────────────────────────────┘    │    │
        │   └─────────────────────────────────────┘    │
        └─────────────────────────────────────────────┘
              di wires concrete adapters → use cases
```

**The Dependency Rule:** `services` depends only on `model` and on repository **interfaces**;
concrete repositories (`repo`) and delivery (`handler`) depend on `services`; `di` is the only place
that knows every concrete type. Keep DB rows and API DTOs out of `model`.

## Project layout

```
neptune-pamm-server/
├── cmd/
│   └── main.go              # entrypoint: load config, build container (di), start server
├── internal/                # private application code (not importable by other modules)
│   ├── model/               # domain entities — pure business objects, no external deps
│   ├── services/            # use cases — allocation, rollover, fees, accounts (+ port interfaces)
│   ├── repo/                # repository implementations (Postgres/Mongo/Redis) behind interfaces
│   ├── handler/             # delivery adapters — HTTP/REST handlers, request/response mapping
│   ├── proto/               # protobuf / wire contracts (inter-service, NATS payloads)
│   ├── db/                  # database driver wiring (connections, migrations, pools)
│   └── di/                  # dependency-injection composition root
├── pkg/                     # reusable, project-agnostic libraries
│   ├── config/              # env/config loading
│   ├── logger/              # structured logging
│   └── utils/               # shared helpers
├── Dockerfile
├── Makefile
├── .env                     # local configuration (not committed)
└── README.md
```

## Tech stack

| Concern | Choice |
|---|---|
| Language | Go |
| Transactional DB / ledger | PostgreSQL |
| Events, time-series, statements | MongoDB |
| Hot state, sessions, cache | Redis |
| Message broker | NATS (JetStream) |
| Realtime to UIs | WebSocket |
| Platform interconnect (MT5) | Webhook + WebSocket (external connector) |
| Packaging | Docker |

## Getting started

> **Status:** project scaffold. The directory structure and tooling are in place; service code is
> being filled in layer by layer (domain → use cases → adapters).

```bash
# clone
git clone git@github.com:ratheeshkumar25/neptune-pamm-server.git
cd neptune-pamm-server

# configure
cp .env.example .env        # then edit values

# common tasks (see Makefile)
make run                    # run the server locally
make build                  # compile the binary
make test                   # run tests
make docker                 # build the container image
```

## Design references

The full design — architecture, data model, allocation/fee math, MT5 integration, API parity, and
the phased development plan — lives in the sibling [`../neptune/`](../neptune/README.md) docs.
