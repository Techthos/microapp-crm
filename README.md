# microapp-crm

A local-first, single-user sales CRM in one self-contained Go binary. Capture **Leads**, convert the
good ones into **Contacts**, and track opportunities as **Deals** through a pipeline — all stored in
a single embedded [bbolt](https://github.com/etcd-io/bbolt) file, no server and no network.

The same data is reachable through two surfaces, selected at launch:

- a **tview** terminal UI for the human operator, and
- an **MCP stdio server** so an AI assistant can read and update the CRM.

The two modes run one at a time (bbolt is single-writer). The full design lives in
[`docs/SPECIFICATIONS.md`](docs/SPECIFICATIONS.md) — the source of truth for what this app is.

## Getting Started

Requires Go 1.26+. Two dev tools are used for formatting and linting:

```sh
go install mvdan.cc/gofumpt@latest
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

### Build, run, test, lint

```sh
make build        # go build ./...
make run          # go run . (defaults to -mode tui)
go run . -mode mcp  # start the MCP stdio server
make test         # go test ./... -race -cover
make fmt          # gofumpt -w .
make lint         # golangci-lint run
make check        # fmt + tidy + lint + test
```

> The surfaces are not implemented yet — this is the initial scaffold. `make build` / `make test`
> are green from day one; running a mode currently reports "not yet implemented".
