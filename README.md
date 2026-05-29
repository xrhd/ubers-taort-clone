# Uber TAROT Replica

A replica of Uber's **TAROT (TARgeting and Observation Toolkit)** — an incentive targeting and optimization system — built with [encore.dev](https://encore.dev) and Go.

TAROT optimally allocates incentives (promotions, discounts, rewards) to users under budget constraints by modeling the problem as a **Multiple Knapsack Problem**.

## Architecture

### System Overview

```mermaid
graph LR
    BA["Business Associates"] -->|Configure Incentives| ConfigUI

    subgraph ConfigUI["Configuration UI Tool<br/><i>ReactJS + NodeJS</i>"]
    end

    subgraph Cron["Cron<br/><i>Golang + Cadence + Cassandra</i>"]
    end

    ConfigUI -->|"Incentives CRUD<br/><i>protobuf over gRPC</i>"| Orchestrator
    Cron -->|"Periodic Trigger<br/><i>kafka</i>"| Orchestrator

    subgraph Orchestrator["Orchestrator<br/><i>Golang Application</i>"]
    end

    Orchestrator -->|"Usecase Store"| DB[("Docstore")]

    Orchestrator -->|"Get users from cohorts<br/><i>protobuf over gRPC</i>"| Segmentation
    Orchestrator -->|"GetPredictions<br/><i>kafka</i>"| MLGateway
    Orchestrator -->|"Get Configured Budgets<br/><i>protobuf over gRPC</i>"| Budgeting
    Orchestrator -->|"Get Remaining Budgets<br/><i>callstack</i>"| Pacer
    Orchestrator -->|"Get Optimized Investments<br/><i>kafka</i>"| Optimizer
    Orchestrator -->|"Assign Incentives<br/><i>kafka</i>"| Domain

    subgraph Segmentation["Segmentation Platform<br/><i>Spark + Blob Storage</i>"]
    end

    subgraph MLGateway["ML Gateway<br/><i>Spark + Cassandra</i>"]
    end

    subgraph Budgeting["Budgeting Platform<br/><i>Golang Application</i>"]
    end

    subgraph Pacer["Budget Pacer<br/><i>Golang Library</i>"]
    end

    subgraph Optimizer["Multi-lever Optimizer<br/><i>Ray + Uniflow</i>"]
    end

    subgraph Domain["Domain Services<br/><i>Golang Application</i>"]
    end
```

### Targeting Run Flow

```mermaid
sequenceDiagram
    participant Cron
    participant Orchestrator
    participant Segmentation as Segmentation Platform
    participant MLGateway as ML Gateway
    participant Pacer
    participant Optimizer
    participant Domain as Domain Services

    Cron->>Orchestrator: Trigger targeting run
    Orchestrator->>Segmentation: Get all users
    Orchestrator->>MLGateway: Get predictions
    Orchestrator->>Pacer: Get all remaining budgets
    Orchestrator->>Optimizer: Run optimizer for all opportunities
    Orchestrator->>Domain: Assign incentives
```

### Services

| Service | Description |
|---|---|
| `orchestrator` | Central coordinator — triggers and manages targeting runs |
| `segmentation` | Defines and resolves user cohorts |
| `mlgateway` | Serves ML predictions for incentive response |
| `budgeting` | Manages configured budgets per program |
| `pacer` | Tracks remaining budgets and pacing |
| `optimizer` | Solves the multiple knapsack allocation |
| `domain` | Assigns incentives to users |
| `config` | CRUD API for incentive configuration |

## Getting Started

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Encore CLI](https://encore.dev/docs/go/install) — `brew install encoredev/tap/encore`
- [Docker](https://docker.com) (for local databases)

### Run locally

```bash
encore run
```

Open the local dev dashboard at http://localhost:9400/

### Run tests

```bash
encore test ./...
```

## References

- [TAROT Paper (arXiv)](https://arxiv.org/abs/2407.19078)
- [Solving Multiple Knapsack at Scale (Uber Blog)](https://www.uber.com/br/pt-br/blog/solving-multiple-knapsack/)
- [Encore.dev Go Docs](https://encore.dev/docs/go)
- [Encore Examples](https://github.com/encoredev/examples)
