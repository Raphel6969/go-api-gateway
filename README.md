# 🚀 Production-Grade API Gateway in Go

A high-performance, resilient, and extensible API Gateway engineered completely from scratch in Go. Designed with zero-trust security, distributed resiliency patterns, and cloud-native operational controls.

---

## 🏗️ Architecture & Core Features

* **Reverse Proxy Engine:** Custom path rewriting and upstream forwarding.
* **Onion-Model Middleware Chain:** Composable, pluggable request-processing pipeline.
* **Zero-Trust Security:** Multi-tenant API key validation and JWT authentication with automatic `X-User-ID` header injection.
* **Rate Limiting:** Thread-safe token bucket rate limiter preventing upstream abuse.
* **Load Balancing:** Atomic round-robin load distribution across target instances.
* **Active Health Checking:** Background background checks with automatic faulty-instance ejection.
* **Resilient Transports:** Exponential backoff retries with idempotent safety checks and state-machine circuit breakers.
* **Caching Engine:** In-memory cache-aside architecture with user-segmented cache keys.
* **Dynamic Configuration:** Decoupled runtime settings managed via YAML configuration files with zero-downtime hot-reloading (`fsnotify` + atomic memory swaps).
* **Developer CLI:** Built with **Cobra**, offering clean commands (`start`, `validate`) mimicking enterprise tooling like NGINX and Kubernetes.
* **Containerized Deployment:** Multi-stage Docker build optimized for minimal footprint and secure containerization via Docker Compose.

---

## 📁 Project Structure

```text
api-gateway/
├── cmd/
│   └── gateway/         # CLI commands (root, start, validate)
├── internal/
│   ├── auth/            # JWT & API Key authenticators
│   ├── cache/           # In-memory cache-aside engine
│   ├── config/          # YAML config parser engine
│   ├── health/          # Background health check routines
│   ├── loadbalancer/    # Round-robin load balancing algorithms
│   ├── middleware/      # Onion-model middleware implementations
│   ├── proxy/           # Reverse proxy engine & transport resilience
│   ├── ratelimit/       # Thread-safe token bucket rate limiter
│   └── router/          # Dynamic path routing engine
├── pkg/
│   └── logger/          # Structured JSON logging framework
├── gateway.yaml         # Dynamic infrastructure configuration
├── Dockerfile           # Multi-stage production container build
└── docker-compose.yml   # Multi-backend containerized orchestration