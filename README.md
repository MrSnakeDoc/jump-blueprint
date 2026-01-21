# Jump 🚀

> **The missing link between your browser and your self-hosted services**

Jump is an intelligent HTTP redirection service that turns your browser's address bar into a powerful launcher for your self-hosted infrastructure. Type a fuzzy keyword, get instantly redirected to the right service.

## Why Jump?

### The Problem

You self-host dozens of services: Jellyfin, Nextcloud, Grafana, Traefik, etc. To access them, you either:

1. **Type the full URL** → `https://jellyfin.long-domain.example.com` (tedious)
2. **Use bookmarks** → cluttered, not keyboard-driven, requires manual sync
3. **Click through Homepage dashboard** → requires mouse, slow, breaks flow

**None of these are instant. None are fuzzy. None learn from your behavior.**

### The Solution

Jump turns your browser into a fuzzy launcher. Configure one shortcut (`jp`) and type partial service names to instantly redirect. See [how to use it](#4-use-it).

**No configuration files to maintain** — Reuses your existing Homepage `services.yaml`  
**No manual mappings** — Fuzzy matching figures out what you mean  
**Learns from you** — More usage = higher ranking in ambiguous queries

---

## Features

### 🎯 Smart Fuzzy Matching

Type partial service names:
- `je` → matches `jellyfin`, `jellyseerr`
- `trae` → matches `traefik`
- `prx` → exact match priority for `prx.example.com` over `prometheus.example.com`

Supports subdomain matching: `jelly.prod` matches `jellyfin.production.example.com`

### 🧠 Usage Learning

Redis-backed learning system improves accuracy over time:
- Frequently accessed services get priority
- Cache hits skip validation (instant redirects)
- Logarithmic scoring prevents dominance

### 🔒 Security First

- **No open redirects**: Domain whitelist enforcement (`JUMP_ALLOWED_HOSTS`)
- **TLS validation**: Only redirects to services with valid HTTPS certificates
- **IP restrictions**: Protect admin endpoints with CIDR allowlists
- **Host validation**: Prevent DNS rebinding attacks

### ⚡ Performance

- **Instant redirects**: Cache hits serve in <10ms
- **Parallel validation**: Top N candidates validated concurrently
- **Configurable candidates**: Tune `MaxCandidates` for speed vs accuracy
- **Long TTL**: Cached resolutions valid for hours

### 🛠️ Zero Config Overhead

- **Single source of truth**: Reuses Homepage `services.yaml`
- **Auto-reload**: Manual trigger via `/reload` endpoint
- **No discovery**: DNS validates, doesn't enumerate
- **Stateless design**: Redis is optimization, not requirement

### 📊 Reporting

- **Health checks**: `/healthz`, `/readyz` for Kubernetes/Docker probes
- **Infrastructure status**: `/infra` JSON endpoint shows system state
- **Structured logging**: JSON logs for production, colored for dev
- **Graceful shutdown**: Proper cleanup of connections and goroutines

### � External Bookmarks

Quick access to external URLs (not part of your self-hosted services) with fuzzy matching:
- Prefix queries with `@` to search bookmarks only
- Reuses Homepage `bookmarks.yaml` (optional)
- No TLS validation or domain restrictions for external URLs
- Examples: `jp @chat` → ChatGPT, `jp @hub` → Docker Hub

### �🔧 Internal Shortcuts

Quick access to Jump's own endpoints with fuzzy matching:
- `jp /inf` → `/infra` (infrastructure status)
- `jp /hea` → `/healthz` (health check)
- `jp /rea` → `/readyz` (readiness probe)

---
## Search Routing (Realms)

Jump uses **prefixes to isolate search spaces**. Each prefix targets a specific "realm" and searches are **strictly separated**—typing `@chat` will never match your self-hosted services.

### Query Prefixes

| Prefix | Realm | Description | Example |
|--------|-------|-------------|---------|
| **(none)** | **Services** | Self-hosted services from `services.yaml` | `jp jelly` → Jellyfin |
| `.` | **Subdomains** | Explicit subdomain matching for services | `jp jelly.prod` → jellyfin.production.example.com |
| `/` | **Internal** | Jump's own endpoints (health, infra, reload) | `jp /inf` → /infra |
| `@` | **Bookmarks** | External URLs from `bookmarks.yaml` | `jp @chat` → ChatGPT |

### Key Behaviors

- **Isolated searches**: `@` queries only search bookmarks, never services
- **No fallback between realms**: If no bookmark matches `@chat`, you get redirected to Homepage—not to a service
- **Explicit routing**: Use `.` to disambiguate subdomain matches (e.g., `jelly.home` vs just `jelly`)
- **Fast internal access**: `/` prefix gives instant access to Jump's admin endpoints

**Example:** `jp jelly` searches services, `jp @jelly` searches bookmarks—two completely separate result sets.

---
## Quick Start

### 1. Prerequisites

- **Go 1.21+** (for building from source)
- **Redis 7+** (for caching and usage learning)
- **Homepage** (or any `services.yaml` compatible file)

### 2. Installation

```bash
# Clone the repository
git clone https://github.com/MrSnakeDoc/jump.git
cd jump

# Copy and configure environment
cp .env.example .env
nano .env  # Edit with your settings

# Build
make build

# Run
./bin/jump
```

### 3. Configure Your Browser

Add a custom search engine:

**Chrome/Brave/Edge/Firefox:**
1. Settings → Search Engines → Manage search engines
2. Add new search engine:
   - **Name**: Jump
   - **Keyword**: `jp`
   - **URL**: `https://jump.example.com/search?q=%s`

### 4. Use It!

In your browser's address bar, type `jp` followed by your query.
You can type multiple words; Jump will fuzzy match them against your services.
```
jp jel fin      → Redirect to Jellyfin (https://jellyfin.example.com)
jp jel se       → Redirect to Jellyseerr (https://jellyseerr.example.com)
jp ne clo       → Redirect to Nextcloud (https://nextcloud.example.com)
jp graf         → Redirect to Grafana (https://grafana.example.com)
jp adg          → Redirect to AdGuard (https://adguard.example.com)
jp adg ha       → Redirect to AdGuard second instance (https://adguardha.example.com)
jp adg.home     → Redirect to adGuard.home (https://adguard.home.example.com)
jp prx          → Redirect to Proxmox (https://proxmox.example.com)
jp @chat        → Redirect to ChatGPT (https://chat.openai.com/)
jp @hub         → Redirect to Docker Hub (https://hub.docker.com/)
jp @cl fla      → Redirect to Cloudflare (https://dash.cloudflare.com/)
jp /inf         → View Jump's infrastructure status (https://jump.example.com/infra)
```

Using multiple words helps disambiguate similar services. For example, `jp jel se` is more likely to match `jellyseerr` than just `jp jel`.

---

## Configuration

Jump is configured entirely through environment variables. All settings are documented in [`.env.example`](.env.example).

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `JUMP_SERVICE_FILE` | Path to Homepage services.yaml | `/app/services.yaml` |
| `JUMP_BOOKMARK_FILE` | Path to Homepage bookmarks.yaml (optional) | `/app/bookmarks.yaml` |
| `JUMP_HOMEPAGE_URL` | Fallback URL when no match found | `https://homepage.example.com` |
| `JUMP_ALLOWED_HOSTS` | Comma-separated allowed Host headers | `jump.example.com,*.example.com` |
| `JUMP_REDIS_ADDR` | Redis server address | `localhost:6379` |
| `JUMP_REDIS_DB` | Redis database number | `0` |

### Optional Variables

#### Logging

| Variable | Default | Description |
|----------|---------|-------------|
| `JUMP_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `JUMP_PRETTY_LOG` | `true` | Colored console logs (false for JSON) |

#### Server

| Variable | Default | Description |
|----------|---------|-------------|
| `JUMP_LISTEN_PORT` | `:8080` | Server listen address |
| `JUMP_SHUTDOWN_TIMEOUT` | `5s` | Graceful shutdown timeout |

#### Redis Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `JUMP_REDIS_USERNAME` | `""` | Redis username (optional) |
| `JUMP_REDIS_PASSWORD` | `""` | Redis password (optional) |
| `JUMP_REDIS_PASSWORD_REQUIRED` | `true` | Require password to be set |

#### Performance Tuning

| Variable | Default | Description |
|----------|---------|-------------|
| `JUMP_TLS_TIMEOUT` | `500ms` | Timeout for TLS validation per service |
| `JUMP_MAX_CANDIDATES` | `3` | Max candidates to validate (0 = unlimited) |
| `JUMP_RELOAD_INTERVAL` | `24h` | Auto-reload services.yaml interval |
| `JUMP_SKIP_TLS_VALIDATION` | `false` | Skip TLS checks (dev only) |

#### Security

| Variable | Default | Description |
|----------|---------|-------------|
| `JUMP_ALLOWED_CIDRS` | `""` | IP ranges for admin endpoints (CIDR) |
| `JUMP_TRUST_PROXY` | `true` | Trust X-Forwarded-For headers |

---

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/search?q=<query>` | GET | Main search endpoint. Fuzzy matches query and redirects to service. |
| `/healthz` | GET | Liveness probe. Returns `{"status": "ok"}` |
| `/readyz` | GET | Readiness probe. Validates Redis connection. |
| `/infra` | GET | System status (protected). Shows routing mode and component health. |
| `/reload` | POST | Manual services.yaml reload (protected). Returns 202 on success. |

---

## Architecture

Jump follows **clean architecture** principles with clear separation of concerns (non exhaustive):

```
cmd/jump/                    → Application entry point
internal/
  ├── app/                   → Application lifecycle and dependency wiring
  ├── config/                → Environment variable management
  ├── domain/                → Core business logic
  │   ├── resolver.go        → Query parsing and service matching
  │   ├── scoring.go         → Fuzzy matching algorithm and ranking
  │   ├── service.go         → Service domain model
  │   ├── bookmark.go        → Bookmark domain model
  │   ├── bookmark_scoring.go → Bookmark fuzzy matching
  │   └── status.go          → TLS validation logic
  ├── httpserver/            → HTTP layer
  │   ├── handlers/          → Request handlers (search, health, reload)
  │   ├── mw/                → Middleware (CORS, rate limit, IP filter)
  │   └── routes/            → Route registration (registry pattern)
  ├── index/                 → In-memory service index
  ├── logger/                → Structured logging (zap)
  ├── redis/                 → Redis connection with retry logic
  ├── scheduler/             → Background jobs
  │   ├── homepage_reload.go → Periodic services.yaml reload
  │   ├── bookmark_reload.go → Periodic bookmarks.yaml reload
  │   ├── garbage_collector.go → Cleanup disabled services/bookmarks
  │   └── redis_sync.go      → Sync usage counters from Redis
  ├── sources/               → Service file parsers
  │   └── homepage/          → Homepage YAML parser and mapper
  │       ├── loader.go      → Services YAML loader
  │       ├── bookmark_loader.go → Bookmarks YAML loader
  │       └── mapper.go      → Domain mappers
  ├── store/redis/           → Redis persistence layer
  │   ├── cache.go           → Query result caching
  │   ├── usage.go           → Usage counter tracking
  │   └── service.go         → Service metadata storage
  └── utils/                 → Pure utility functions
```

**Core Principles:** Jump validates services via TLS handshakes (no DNS enumeration for security). Redis provides caching and learning but isn't required—degraded mode works without it. Every request is stateless for horizontal scaling. Failed matches redirect to Homepage instead of 404.

---

## Deployment

### Docker

```bash
docker run -d \
  --name jump \
  -p 8080:8080 \
  -v /path/to/services.yaml:/app/services.yaml:ro \
  --env-file .env \
  ghcr.io/mrsnake/jump:latest
```

### Docker Compose

refer to [compose.example.yml](compose.example.yml) for a full example including Traefik integration.

### Behind Traefik

```yaml
# docker-compose.yml labels
labels:
    - 'traefik.enable=true'
    - 'traefik.http.routers.jump.entrypoints=http,https'
    - 'traefik.http.routers.jump.rule=Host(`jump.your-domain.ext`)'  # Replace with your domain
    - 'traefik.http.services.jump.loadbalancer.server.port=8080'
    - 'traefik.http.routers.jump.service=jump'
    - 'traefik.http.routers.jump.tls=true'
    - 'traefik.http.routers.jump.tls.certresolver=<your-cert-resolver-name>'
```

---

## Development

### Local Setup

1. **Start Redis:**
   ```bash
   make start-redis
   ```

2. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your local settings
   ```

3. **Run the app:**
   ```bash
   make start
   ```

**Available Commands:** See [Makefile](Makefile) for all targets (`build`, `test`, `lint`, `vuln`, `start-redis`, etc.)

---

## How It Works

Jump parses Homepage's `services.yaml` (and optionally `bookmarks.yaml`) on startup (and every 24h or via `/reload`). 

**For services** (`jp jelly`): Checks Redis cache first. On miss, fuzzy-matches all services with a scoring algorithm: exact match (300pts), prefix (75pts), substring (50pts), fuzzy (25pts), plus usage learning (logarithmic boost). Top candidates are TLS-validated in parallel, first success wins. Results cache for 6h.

**For bookmarks** (`jp @chat`): Fuzzy-matches external bookmarks (no cache, no TLS validation, no domain restrictions). Directly redirects to the best match.

**For internal endpoints** (`jp /inf`): Fuzzy-matches internal Jump endpoints.

Failed matches redirect to Homepage—no 404s.

---

## Roadmap

**v0.1.x - Production Ready** ✅
- [x] Configuration management
- [x] Redis connection with retry logic
- [x] HTTP server with middleware
- [x] Health checks and monitoring
- [x] Infrastructure status endpoint
- [x] Homepage YAML parser
- [x] In-memory service indexing
- [x] Fuzzy matching resolver
- [x] Usage scoring system
- [x] TLS validation
- [x] Search endpoint
- [x] Manual reload endpoint
- [x] Internal shortcuts (`/inf`, `/hea`)
- [x] External bookmarks support (`@` prefix)
- [x] Garbage collector for disabled services/bookmarks

**v1.1 - Enhancements** 🚧
- [ ] Multi-source support (not just Homepage)

---

## License

MIT License — See [LICENSE](LICENSE) file for details.

---

## Disclaimer

**This project is:**

- ✋ **Opinionated** — Built for my workflow but can be adapted for yours freely with a fork
- 🏠 **Personal** — Designed for my infrastructure
- 📎 **Tightly coupled** — Homepage YAML format only
- 🚫 **Not general-purpose** — Fork and adapt it to your needs

**This project is stable and feature-complete for my needs.**

I’m not actively accepting feature requests or pull requests.  
Forks are encouraged if you want to adapt it to your own workflow.

**Security issues are welcome and will be handled responsibly.**

🔒 **Security issues are welcome** — Please report responsibly

If you want to use it: **fork it** and adapt it to your own needs.

---

## Acknowledgments

- **[Homepage](https://github.com/gethomepage/homepage)** - For the amazing dashboard that inspired and power this project
- **[Traefik](https://traefik.io/)** - For making reverse proxy configuration painless
- **Community** - For feedback and encouragement

---

**Built with ❤️ for self-hosters**

*Jump is designed for self-hosted environments behind trusted reverse proxies. Ensure proper TLS termination and access controls are configured at the proxy level.*
