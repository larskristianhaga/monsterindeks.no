# Monsterindeks.no

This repository contains the sourcecode for the website [monsterindeks.no](https://www.monsterindeks.no/)

A simple webpage that tracks the price of Monster Energy on a daily basis, and displays the data in a graph.

Data is inserted from a cronjob that runs every day at 12:00, and is hosted on [cron-job.org](https://console.cron-job.org).

## Running locally

```bash
DATABASE_PATH=./monsterdatabase.db go run .
```

Or with Docker:

```bash
docker run -it --rm -p 8080:8080 -v "$PWD/data:/data" $(docker build -q .)
```

The SQLite database and the `monsters` table are created on startup if they don't exist. Visit `/insert-latest-monster-price` to add today's price.

To fill a local database with ~180 days of mock prices (replaces any existing rows):

```bash
sqlite3 monsterdatabase.db < scripts/seed.sql
```

## Hosting

The app runs on [Fly.io](https://fly.io) with the SQLite database on a volume mounted at `/data`. Volumes are local to one machine, so the app must run on exactly one machine:

```bash
fly scale count 1
```