# Monsterindeks.no

This repository contains the sourcecode for the website [monsterindeks.no](https://www.monsterindeks.no/)

A simple webpage that tracks the price of Monster Energy on a daily basis, and displays the data in a graph.

Data is inserted from a cronjob that runs every day at 12:00, and is hosted on [cron-job.org](https://console.cron-job.org).

## Running locally

```bash
docker run -it --rm -p 8080:8080 $(docker build -q .)
```