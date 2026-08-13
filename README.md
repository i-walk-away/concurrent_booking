# Concurrent Booking

A high-concurrency backend service for cinema seat reservations, designed to handle concurrent attempts to book the same
seat safely.

## Features

* Concurrency test with 100.000 simultaneous booking attempts for the same seat
* Temporary seat holds with a 2 minute TTL
* Booking confirmation and release
* Atomic seat reservation using Redis `SET NX`

## Run

Start redis container:

```bash
docker compose up --build
```

Run the application:

```bash
go run ./cmd
```

Simple HTML frontend will be available at:

```text
http://localhost:8080
```

## Tests

Run all tests:

```bash
go test ./...
```

Run with the race detector:

```bash
go test -race ./...
```

