# Go Wallet Service

[🇷🇺 README на русском](https://github.com/uiNoiSE/go-wallet/blob/main/README.md)

A highly concurrent, fault-tolerant RESTful wallet service designed to handle heavy loads under strict concurrency guarantees. Built with a focus on fast writes and data integrity, reaching **3600+ RPS** per single wallet without database deadlocks or data races.

### 🚀 Architecture Decisions (Why Append-Only?)

The core requirement of this task was handling heavy concurrent traffic (**1000+ RPS per single wallet**) without dropping requests ($50X$ errors) or corrupting balances.

#### The Problem with Naive Approaches
Typically, developers use `SELECT FOR UPDATE` (Row-level locking) to handle balance updates. Under high concurrency on a single wallet, this leads to:
1. **Database Deadlocks:** Threads blocking each other while waiting for the lock.
2. **Performance Bottlenecks:** Requests line up sequentially, causing pool exhaustion and context deadline timeouts ($504$ Gateway Timeouts).

#### Our Solution: Append-Only Transaction Log (Event Sourcing Light)
Instead of mutating a single balance row in place, this service treats the database as an **Append-Only immutable ledger**.
* **Deposits & Withdrawals:** Every request is a pure `INSERT` into the `transactions` table. Inserts do not lock existing rows, allowing Postgres to execute them completely in parallel at maximum hardware capability.
* **Balance Calculation:** The current balance is computed dynamically via an optimized aggregation (`SUM(amount)`). 
* **Overdraft Protection (In-Flight Checks):** For withdrawals, the service uses a transient snapshot isolation check. If concurrent double-spending attempts occur, the state remains consistent because money cannot be spent from an uncommitted ledger state.

This paradigm shifted the system bottleneck from slow, synchronous DB row-locks to blazing-fast parallel disk writes.

### 📊 Performance & Load Testing

The system was stress-tested under concurrent load simulating race conditions on a single account.

#### Concurrent Test Results:
* **Total Requests:** 500 parallel transactions
* **Execution Time:** ~138.6 ms
* **Throughput (RPS):** **~3607 req/sec**
* **Data Integrity:** 100% accurate. Final balance matched the atomic invariant exactly. Zero failed requests, zero 50X errors.

### 🛠️ Tech Stack

* **Language:** Go (1.26)
* **Database:** PostgreSQL (18)
* **Migration Tool:** Goose
* **Containerization:** Docker / Docker Compose

### ⚡ Quick Start (Docker Compose)

> [!NOTE]
> For your convenience, the `example.config.env` file intentionally contains fully pre-configured and production-ready environment variables.

The entire system—including the application, database, and migrations—is configured to launch seamlessly with a single command.

```bash
docker compose up --build
```

### 📡 API Specification
#### 1. Create Wallet
Creates a new wallet instance with a 0.00 default balance.

* **URL**: `/api/v1/wallet`
* **Method**: `POST`
* **Response (201 Created):**
```json
{
  "id": "eb7dc3a2-ff82-4c83-98d7-4431ee9ac0ab",
}
```

#### 2. Process Transaction
Executes an atomic balance operation.
* **URL**: `/api/v1/wallet`
* **Method**: `POST`
* **Тело запроса:**
```json
{
  "id": "eb7dc3a2-ff82-4c83-98d7-4431ee9ac0ab",
  "operationType": "DEPOSIT",
  "amount": 1000.50
}
```
(Supported types: `DEPOSIT`, `WITHDRAW`)
* **Response (200 OK)**

#### 3. Get Wallet Balance
Fetches the aggregated current balance.
* **URL**: `/api/v1/wallet/:id`
* **Method**: `GET`
* **Response (200 OK):**
```json
{
  "balance": 1000.50
}
```

### 🧪 Running Tests Locally
```bash
go test -v ./...
```
