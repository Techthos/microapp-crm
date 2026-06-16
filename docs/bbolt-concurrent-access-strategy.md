# bbolt concurrent-access strategy: connection-per-operation

A portable strategy for letting **two or more independent processes** read and write the same
[bbolt](https://github.com/etcd-io/bbolt) file, without the "open at startup, hold the handle for
the process lifetime" model that makes bbolt single-process. Application-agnostic; copy it anywhere.

> **TL;DR** — bbolt's file lock is taken at `Open`, not per-transaction, and a held *read* lock
> blocks *all* writers. The only way two processes can both write is for **neither to hold any handle
> while idle**: open for one operation, then close. Wrap the open in a timeout + backoff retry so a
> collision becomes a sub-second wait. Detect external writes by polling bbolt's monotonic txid.

## 1. Why the default model is single-process

bbolt's lock is an OS file lock (`flock` on Unix, `LockFileEx` on Windows). Two facts decide
everything:

1. **The lock is acquired at `Open`, not per transaction.** There is no "lock only while writing";
   opening read-write locks the file until `Close`. The usual `bolt.Open(...); defer db.Close()`
   therefore holds a process-wide lock for the whole process lifetime. A second process that opens
   read-write blocks until its `Timeout`, then fails with `ErrTimeout`.
2. **Two lock modes conflict in the direction that bites:** read-write takes `LOCK_EX` (exclusive);
   `ReadOnly: true` takes `LOCK_SH` (shared). `LOCK_EX` cannot be granted while **any** `LOCK_SH` is
   held — even one held by the same process on a different descriptor.

So the tempting design — keep a read-only handle open for fast reads, open a second RW handle only
to write — **deadlocks**: the persistent `LOCK_SH` blocks the writer's `LOCK_EX`, including its own.
Two long-lived processes each holding an idle read handle means *no write ever succeeds*.

**Consequence:** to let two processes both write, *no process may hold any handle while idle.*

## 2. The algorithm: connection-per-operation

Open the database for the duration of **one operation** and close it immediately. Idle processes
hold no lock, so any process is free to grab the (exclusive, brief) write lock when it needs it.
Reads and writes stay serialized at the file level, but each lock lasts milliseconds — for
low-contention workloads it behaves as if parallel.

```
idle      → no handle, no lock            (any process may write)
read op   → Open(ReadOnly) → View   → Close   (LOCK_SH, milliseconds)
write op  → Open(RW)       → Update → Close   (LOCK_EX, milliseconds)
```

Three rules make it correct and robust:

- **Per-open `Timeout` + backoff retry.** Keep the per-attempt timeout short (~75ms) and retry with
  growing backoff up to a total budget (~3s). On collision the loser waits, not fails. Retry **only**
  on `bolt.ErrTimeout`; any other error is fatal. Because retry uses `time.Sleep` (blocking), never
  call an operation on a UI/event-loop goroutine — run it on a worker and hand the result back.
- **One-time bootstrap.** `ReadOnly` opens require the file and buckets to already exist (bbolt
  can't create a file read-only). At startup, open read-write **once**, run an idempotent migration
  (`CreateBucketIfNotExists` for every bucket), and close. Every later operation assumes the schema.
- **Each operation is its own transaction.** Keep every cross-entity, must-be-atomic use-case inside
  a **single** `update(func(tx){...})` so it commits or rolls back as a unit — you can no longer span
  multiple `view`/`update` calls in one transaction.

The `Store` holds only the **path**, never a live `*bolt.DB`. An `open(readOnly bool)` helper runs
the timeout/retry loop; thin `view`/`update` wrappers call `open`, run the `bolt.Tx` function, and
`defer db.Close()`. (Reference Go implementation: see the bbolt GoDoc for `Options.Timeout`,
`Options.ReadOnly`; the wrapper is ~40 lines.)

## 3. Cross-process freshness: detecting "someone else wrote"

The data layer is never stale — every read opens fresh and sees the latest committed state. But a
**long-lived UI** caches a *rendered snapshot* in its widgets that must refresh when another process
writes.

bbolt stamps a **monotonically increasing transaction ID** in its meta page on every committed
write; a read transaction's `tx.ID()` returns the latest committed ID. Comparing it across reads is
a near-free "did anything change?" probe (one `Open` + empty `View`, no data scan). Wire it into a
long-lived reader as a background poll:

```
every 1–2s on a worker goroutine:
    now := store.txid()            // one Open + empty View
    if now != lastSeen:
        data := store.fetchAll()             // re-query off the UI thread
        queueUIUpdate(func(){ render(data) }) // tiny mutation on the UI thread
        lastSeen = now
```

Pair it with a **manual refresh key**. Notes: a no-op write still bumps the txid (occasional
identical re-render — harmless); `tx.ID()` is authoritative while file `mtime`/size are not (bbolt
writes via `mmap` + `fsync`); `fsnotify` works too but emits several events per commit and still
needs a txid confirm, so the poll is simpler.

## 4. Trade-offs & decision guide

Connection-per-operation costs a per-op `Open` (mmap + meta read, a few ms), a cold page cache each
op, and lower write throughput (one fsync + open per op) — in exchange for multiple writer processes
and always-fresh cross-process reads. **Read-your-own-writes across processes is eventual**: between
two operations another process may have written (identical to alt-tabbing between two windows).

```
Do multiple OS processes need to write the same bbolt file?
├─ No → use the standard single-handle model. This doc doesn't apply.
└─ Yes
   ├─ Can you instead run ONE process hosting all surfaces?
   │     → Prefer that. One handle, no lock dance, full cache & throughput.
   ├─ Low-contention & low-throughput (desktop tool, single user, UI + helper)?
   │     → Use connection-per-operation (this doc).
   └─ Write-heavy or genuinely concurrent (many writers, high TPS)?
         → Wrong tool. Use a client/server DB (SQLite+WAL; Postgres for true concurrency).
```

## 5. Correctness checklist

- [ ] `Store` holds the **path**, never a live `*bolt.DB`; no handle/`*bolt.Tx`/tx-scoped slice
      retained past its helper (copy or unmarshal before returning — the handle closes right after).
- [ ] One-time **bootstrap** opens RW, runs the idempotent migration, closes.
- [ ] `open(readOnly)` retries **only** on `bolt.ErrTimeout`, with short per-attempt `Timeout` +
      backoff up to a budget matched to your contention (documented if non-default).
- [ ] All reads via `view` (ReadOnly), all writes via `update` (RW); each atomic use-case is a
      **single** `update` transaction; operations are short (no network/blocking/user I/O inside).
- [ ] `view`/`update` are **never** called on a UI/event-loop goroutine.
- [ ] Long-lived readers refresh via a **txid poll** (`tx.ID()`) plus a **manual refresh** key.

## References

- bbolt README — https://github.com/etcd-io/bbolt/blob/main/README.md
- `Options.Timeout`, `Options.ReadOnly`, `Tx.ID()` — https://pkg.go.dev/go.etcd.io/bbolt#Options
- `flock(2)` lock semantics — https://man7.org/linux/man-pages/man2/flock.2.html
