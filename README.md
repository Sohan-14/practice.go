## Project Setup:

```bash
mkdir practice
cd practice
go mod init example.com/practice
```

This command this create a go.mod file inside the practice folder

> go mod init example.com/practice
>

#### go.mod

```go
module example.com/practice

go 1.27.3

```

First line is the our project module name. and second one is the go version. it is the version installed in our computer system

#### main.go

Create a main.go file inside the folder practice which is our project folder: you can create folder by using GUI (File manager or code editor - VS CODE) or CLI (Terminal Prompt):

```bash
touch main.go # cli command to create a file
```

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /{$}", homeHandler)
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /api/hello", helloHandler)

	addr := ":8080"
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🚀 Server starting on http://localhost%s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// homeHandler returns a welcome message
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to my Go HTTP server!")
	fmt.Fprintln(w, "Try: /health or /api/hello?name=World")
}

// healthHandler returns server health status
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// helloHandler returns a personalized greeting
func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Hello, %s!", name),
	})
}

// writeJSON is a helper to send JSON responses
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
```

#### Run command:

open terminal and run the command

```bash
go run main.go
```

now open another terminal to test the http server: we can also use postman or browser to test

```bash
curl http://localhost:8080/
# Welcome to my Go HTTP server!
# Try: /health or /api/hello?name=World

curl http://localhost:8080/health
# {"status":"ok"}

curl http://localhost:8080/api/hello?name=Alice
# {"message":"Hello, Alice!"}
```

#### **Key Points:**

- **Standard library only** — no external dependencies needed
- **`http.ServeMux`** — Go's built-in router handles request dispatching
- **`/health` endpoint** — useful for monitoring/load balancer checks
- **JSON responses** — with proper `Content-Type` headers

### API Client used: Bruno

---

---

# **How the Go HTTP Server Works — Full Explanation**

This program creates a **web server**. A web server is a program that does one thing in a loop:

- **Waits** for incoming requests (from a browser, `curl`, etc.)
- **Receives** a request like "give me the page at `/health`"
- **Runs** the matching function
- **Sends back** a response
- Goes back to waiting

## **Starting the Server:**

**Step 1: Create the router**

```go
mux := http.NewServeMux()
```

A **mux** (multiplexer) is a router. It's the "host" that looks at each incoming request's URL and decides *which function* should handle it.

**Step 2: Register the routes**

```go
mux.HandleFunc("GET /{$}", homeHandler)
mux.HandleFunc("GET /health", healthHandler)
mux.HandleFunc("GET /api/hello", helloHandler)
```

This says:

| Pattern | Meaning |
| --- | --- |
| `GET /{$}` | When someone visits **exactly** `/` (the home page) with a GET request → run `homeHandler` |
| GET /health | When someone visits `/health` → run `healthHandler` |
| `GET /api/hello` | When someone visits `/api/hello` → run `helloHandler` |

`GET` is the HTTP method — "give me data" (vs. `POST` = "here's data to save"). And we have more methods

**Step 3: Configure the server**

```go
server := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
}
```

- **`:8080`** — the port to listen on. A port is like a door number. `localhost:8080` means "this machine, door 8080".
- **`Handler: mux`** — all requests go through our router.
- **Timeouts** — if a client is too slow (more than 5s to send, 10s to receive), the server cuts the connection. This prevents slow clients from tying up the server forever.

**Step 4: Start listening**

```go
log.Printf("🚀 Server starting on http://localhost%s", addr)
if err := server.ListenAndServe(); err != nil {
    log.Fatal(err)
}
```

**`ListenAndServe()`** is the heart of the program. It:

- Opens port 8080 and starts listening
- Runs **forever**, handling requests one after another
- **Blocks** — code after it never runs unless something goes wrong

**The Handlers — Processing Requests**

Every handler has the same signature:

```go
func handlerName(w http.ResponseWriter, r *http.Request)
```

- **`w` (ResponseWriter)** — where you *write the answer*. Whatever you write here gets sent back to the client.
- **`r` (Request)** — everything the client *sent you*: the URL, method, headers, query parameters, body...

**Home handler**

```go
func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Welcome to my Go HTTP server!")
    fmt.Fprintln(w, "Try: /health or /api/hello?name=World")
}
```

`fmt.Fprintln(w, ...)` means: "print this text... into the response." Plain text goes back to the client.

**Health handler**

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
```

Returns `{"status":"ok"}`. This endpoint exists so monitoring tools can ask *"are you alive?"* — if they get a 200 back, the server is healthy.

`http.StatusOK` is just the number **200** (HTTP status code for "success").

**Hello handler**

```go
func helloHandler(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    if name == "" {
        name = "World"
    }
    writeJSON(w, http.StatusOK, map[string]string{
        "message": fmt.Sprintf("Hello, %s!", name),
    })
}
```

The key line: `r.URL.Query().Get("name")`.

When you request `/api/hello?name=Alice`:

- The part after `?` is the **query string**
- `name=Alice` is a query parameter
- This line extracts the value `"Alice"`
- If no name was given, we default to `"World"`

**The `writeJSON` Helper**

```go
func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}
```

Three things happen **in this exact order** (order matters!):

1. **Set the header** — tell the client "the body is JSON". The client (browser/curl) uses this to know how to display it.
2. **Write the status code** — 200 = success. Must happen before writing the body.
3. **Encode the body** — convert the Go map into JSON text, e.g. `map["status":"ok"]` → `{"status":"ok"}`.

## **Full Request Lifecycle — What Happens Step by Step**

When you type:

```bash
curl http://localhost:8080/api/hello?name=Alice
```

Here's the full journey:

```markdown
1. curl sends an HTTP request:
      GET /api/hello?name=Alice
      (to localhost, port 8080)
          ↓
2. ListenAndServe() receives it
          ↓
3. The mux checks: "GET /api/hello" → matches! → call helloHandler
          ↓
4. helloHandler reads ?name=Alice from the URL
          ↓
5. Builds: {"message": "Hello, Alice!"}
          ↓
6. writeJSON sets headers, status 200, writes JSON body
          ↓
7. Response travels back to curl
          ↓
8. curl prints: {"message":"Hello, Alice!"}
```

**Understanding the Output**

**In your terminal (the server side):**

```markdown
2026/11/03 10:15:23 🚀 Server starting on http://localhost:8080
```

Then nothing — **silence means it's working**, waiting for requests. To stop it, press `Ctrl+C`.

**In curl (the client side):**

```bash
curl http://localhost:8080/
Welcome to my Go HTTP server!
Try: /health or /api/hello?name=World
```

↑ Plain text, because `homeHandler` used `fmt.Fprintln`.

```bash
curl http://localhost:8080/health
{"status":"ok"}
```

↑ JSON, because `healthHandler` used `writeJSON`.

```bash
curl http://localhost:8080/api/hello?name=Go
{"message":"Hello, Go!"}
```

↑ JSON with your name inserted.

**What about a wrong URL?**

```bash
curl http://localhost:8080/whatever
404 page not found
```

The mux has no route for `/whatever`, so Go automatically returns **404** ("not found"). You didn't write that — it's free behavior from the router.

**Recap**

| Concept | Role |
| --- | --- |
| `http.Server` | Listens on a port, accepts connections |
| ServeMux | Routes URLs → handler functions |
| Handler | Your code that builds the response |
| Request | What the client sent |
| ResponseWriter | Where you write the reply |
| Status code | `200` = OK, `404` = not found |
| JSON | Standard data format for APIs |

**The mental model:** the server is an infinite loop of *receive request → route it → run handler → send response*.

---

---

# **Production / Enterprise Problem Sheet**

**Go HTTP Server — Software Engineering Review**

**Context:** The server works. But "works on my machine" ≠ "survives in production." This sheet reframes the same code through engineering concerns: **reliability, operations, security, maintainability, and team scalability** — not just functionality.

**How to use:** Each item lists the **Gap** (what's missing), the **Risk** (what breaks in production), and the **Engineering Task** (what a real team would do).

## **1. Reliability & Availability**

**P1 — No graceful shutdown**

- **Gap:** `Ctrl+C` or a deploy kills the server instantly, mid-request.
- **Risk:** In-flight requests are dropped; clients see connection resets; deploys cause errors.
- **Task:** Handle `SIGINT`/`SIGTERM` with `os/signal` + `server.Shutdown(ctx)` so the server stops accepting new requests but finishes active ones.

**P2 — Health check is too shallow**

- **Gap:** `/health` always returns OK, even if dependencies are broken.
- **Risk:** Load balancer keeps routing traffic to a dead-but-responding instance.
- **Task:** Split into **liveness** ("is the process alive?") and **readiness** ("can it actually serve traffic / reach its dependencies?").

**P3 — Incomplete timeout coverage**

- **Gap:** No `IdleTimeout`, no per-handler context timeouts.
- **Risk:** Slow-loris attacks, resource exhaustion from hung connections.
- **Task:** Set `IdleTimeout`, propagate `context.Context` with deadlines into handlers.

## **2. Configuration & Environment**

**P4 — Hardcoded configuration**

- **Gap:** Port `:8080` and timeouts live in source code.
- **Risk:** Can't change behavior per environment without recompiling; violates 12-factor principles.
- **Task:** Read config from environment variables / config file with validation and sensible defaults.

**P5 — No secret or environment strategy**

- **Gap:** No plan for DB credentials, API keys, etc.
- **Risk:** Secrets accidentally committed or logged.
- **Task:** Define a secret-management approach (env injection, vault); ensure secrets never hit logs.

## **3. Observability**

**P6 — Unstructured logging, no correlation**

- **Gap:** `log.Printf` plain text; no request ID, level, or fields.
- **Risk:** Impossible to trace one request across logs or filter by severity at scale.
- **Task:** Use structured logging (`log/slog`), attach a **request ID** to every log line.

**P7 — No metrics**

- **Gap:** No counters for requests, latency, error rate.
- **Risk:** You can't detect degradation, set alerts, or capacity-plan.
- **Task:** Expose metrics (e.g., Prometheus `/metrics`): request count, p95 latency, error %.

**P8 — No distributed tracing**

- **Gap:** No way to follow a request across services.
- **Risk:** In a microservices world, debugging latency/failures is guesswork.
- **Task:** Integrate OpenTelemetry trace propagation.

## **4. Error Handling & Robustness**

**P9 — Silently ignored errors**

- **Gap:** `json.NewEncoder(w).Encode(data)` return value is discarded.
- **Risk:** Failures invisible; corrupted responses with no log trail.
- **Task:** Handle and log every error; decide behavior per failure type.

**P10 — No panic recovery**

- **Gap:** A panic in one handler can crash the whole process.
- **Risk:** One bug takes down the entire service (availability incident).
- **Task:** Add `recover()` middleware so a single request failing doesn't kill the server.

**P11 — No structured error responses**

- **Gap:** Errors return raw text or nothing consistent.
- **Risk:** Clients can't reliably parse/handle failures; inconsistent API contract.
- **Task:** Centralize error handling → consistent JSON error shape + correct status codes.

## **5. Security**

**P12 — No TLS/HTTPS**

- **Gap:** Traffic is plaintext.
- **Risk:** Eavesdropping, tampering, credential theft.
- **Task:** Serve over TLS (or terminate at a reverse proxy/load balancer).

**P13 — No input validation or size limits**

- **Gap:** Query/body input accepted unchecked; no `MaxBytesReader`.
- **Risk:** Injection, oversized payloads, memory exhaustion.
- **Task:** Validate all input; cap request body size.

**P14 — No rate limiting**

- **Gap:** Any client can hit the server unlimited.
- **Risk:** Abuse, DoS, cascading overload.
- **Task:** Add rate limiting / throttling middleware.

**P15 — No authentication or authorization**

- **Gap:** All endpoints public.
- **Risk:** Anyone can call anything.
- **Task:** Add auth (e.g., tokens) + authorization checks before handlers.

**P16 — Missing security headers & CORS policy**

- **Gap:** No headers, no CORS rules.
- **Risk:** Common web attacks; uncontrolled cross-origin access.
- **Task:** Add security headers (via a middleware like `secure`) and an explicit CORS policy.

## **6. Testing & Quality**

**P17 — Zero tests**

- **Gap:** No unit, handler, or integration tests.
- **Risk:** Regressions ship to production; refactoring is terrifying.
- **Task:** Use `httptest` to test handlers; establish coverage targets.

**P18 — No CI pipeline**

- **Gap:** No automated build/test/lint on change.
- **Risk:** Broken code reaches main; quality depends on memory.
- **Task:** CI that runs `go build`, `go vet`, `staticcheck`, tests, and coverage on every PR.

## **7. Architecture & Maintainability**

**P19 — Single file, single package**

- **Gap:** Everything in `main`.
- **Risk:** Doesn't scale as code or team grows; merge conflicts; no reuse.
- **Task:** Split into packages (e.g., `cmd/`, `internal/handler`, `internal/...`).

**P20 — No separation of concerns**

- **Gap:** Handlers mix routing, logic, and response formatting.
- **Risk:** Logic can't be tested or reused independently.
- **Task:** Layer it: **transport (handlers) → service (logic) → data**.

**P21 — No dependency injection**

- **Gap:** Dependencies hardcoded inside functions.
- **Risk:** Hard to test, swap, or mock.
- **Task:** Inject dependencies (config, clients) so handlers are testable in isolation.

**P22 — No API versioning or contract**

- **Gap:** Endpoints can change freely.
- **Risk:** Breaking existing clients silently.
- **Task:** Version the API (`/api/v1/...`); document the contract (e.g., OpenAPI).

## **8. Deployment & Operations**

**P23 — No reproducible build / container**

- **Gap:** No `Dockerfile`, no pinned dependencies.
- **Risk:** "Works on my machine"; inconsistent environments.
- **Task:** Multi-stage `Dockerfile`; commit `go.mod`/`go.sum`; build reproducibly.

**P24 — No versioning or build metadata**

- **Gap:** No way to tell *which* build is running.
- **Risk:** Can't debug or roll back reliably.
- **Task:** Inject version/commit via `ldflags`; expose at a `/version` endpoint.

**P25 — No deployment or rollback strategy**

- **Gap:** No plan for releasing or reverting.
- **Risk:** Bad deploys cause prolonged outages.
- **Task:** Define rollout strategy (blue/green, canary) and a tested rollback path.

## **9. Performance & Scalability**

**P26 — No load testing or benchmarks**

- **Gap:** Unknown throughput/latency limits.
- **Risk:** Surprises under real traffic.
- **Task:** Load test (`wrk`/`k6`); add `go test -bench` for hot paths.

**P27 — No resource limits**

- **Gap:** No caps on connections/goroutines.
- **Risk:** Unbounded growth → OOM under load.
- **Task:** Set limits and monitor saturation.

## **The Maturity Ladder**

This is really a progression of engineering maturity:

```markdown
Prototype        → It runs                       (what we built)
Production       → It survives failure & load    (P1, P6, P9, P10, P12, P23)
Enterprise       → It scales across teams/time   (P19–P25, CI, contracts)
```

**Suggested Priority (highest risk first)**

- **P1** Graceful shutdown, **P10** panic recovery — *availability*
- **P12** TLS, **P13** input limits — *security*
- **P6** structured logging, **P7** metrics — *you can't fix what you can't see*
- **P17/P18** tests + CI — *everything else depends on safe change*
- **P19–P21** architecture — *before the codebase/team grows*

**The core mindset shift:**

```markdown
Developing asks "does it work?"
Software engineering asks
"does it keep working —
at 3 a.m.,
under load,
across a team of ten,
for three years?"
```
