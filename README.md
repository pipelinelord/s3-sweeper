# S3 Sweeper 🧹

> A high-performance, concurrent CLI tool for cleaning up stale S3 objects.

![Go Version](https://img.shields.io/badge/go-1.25-blue)
![License](https://img.shields.io/badge/license-MIT-green)

**S3 Sweeper** is a DevOps tool written in Go that scans AWS S3 buckets for files older than a specific age and optionally deletes them. It leverages **Go Concurrency patterns (Worker Pools)** to process thousands of objects in parallel, making it significantly faster than sequential scripts.

## 🚀 Key Features

* **Concurrent Scanning:** Uses a configurable worker pool to process objects in parallel.
* **Safety First:** Default `Dry-Run` mode prevents accidental deletions.
* **AWS SDK v2:** Built on the latest AWS standards.
* **Dockerized:** Ready to run in CI/CD pipelines or Kubernetes.

## 🏗 Architecture

The tool implements a Producer-Consumer pattern:
1.  **Producer:** Iterates over S3 pages and pushes object metadata to a `Job` channel.
2.  **Worker Pool:** `N` goroutines concurrently pull jobs, check age criteria, and delete if necessary.
3.  **Collector:** Aggregates statistics (space saved, files deleted) for a final report.

## 🛠 Installation

### Option 1: Docker (Recommended)
```bash
docker run --rm \
  -e AWS_ACCESS_KEY_ID=... \
  -e AWS_SECRET_ACCESS_KEY=... \
  -e AWS_REGION=us-east-1 \
  ghcr.io/yourusername/s3-sweeper:latest scan --bucket my-logs --days 30