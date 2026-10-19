## [v0.1.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Core architecture: IoC container and module primitives

## [v0.5.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Middleware chain, lifecycle hooks and exception filtering

## [v1.0.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Complete framework: server, router, controller, provider

## [v1.1.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Configuration layer and HTTP exception mapping

## [v1.2.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Graceful shutdown and example application

## [v1.3.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Route grouping, parameter binding and CI

## [v1.4.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Unified response encoding and error handling

## [v1.5.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Validator and HealthCheck middleware

## [v1.6.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Recovery middleware with stack trace capture

## [v1.7.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Compression middleware

## [v1.8.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- IP filter middleware with CIDR support

## [v1.9.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Request logger middleware

## [v2.0.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Timeout middleware

## [v2.1.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- WebSocket upgrade middleware

## [v2.2.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- OpenAPI spec serving with bundled Swagger UI

## [v2.3.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Prometheus metrics middleware

## [v2.4.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Cache middleware

## [v2.5.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- CORS middleware

## [v2.6.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- JWT authentication middleware

## [v2.7.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Rate limit middleware

## [v2.8.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Request ID middleware

## [v2.9.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Session and file upload middleware

## [v3.0.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Database integration via GORM

## [v3.1.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Redis integration

## [v3.2.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- Static file serving middleware

## [v3.3.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- gRPC server integration

## [v3.4.0]

### Removed
- `middleware.ValidatorConfig` (with `BindTarget` and `SkipMissing`). It was
  documented as the configuration for a validator middleware, but no validator
  middleware exists and nothing ever read the struct, so setting it silently did
  nothing. `ValidateRequest(ctx, rules)` remains the supported API. Route-scoped
  validation with struct binding is not implemented

### Added
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Fixed
- API versioning with header, path and query strategies

## [v3.4.1]

77 defects fixed across 34 commits. Grouped by what they were, not by when they
landed; every entry names the file so it can be reviewed directly.

### Security

- **Static file serving escaped its root.** The containment check compared
  `HasPrefix(absPath, absDir)`, so a sibling directory sharing the root's name
  was served: root `/srv/www` served `/srv/www-secret/...`
  (`middleware/static.go`)
- **Static file serving followed symlinks out of the root.**
  `Config.FollowSymLinks` was documented but never read, so with it false a
  symlink inside the root pointing outside it was still served. The resolved
  target is now validated against the resolved root
- **File upload escaped its storage root.** The client-supplied filename was
  joined onto `BaseDir` unsanitised, so `../../etc/cron.d/x` wrote outside it.
  `LocalStorage.Delete` had the same gap (`middleware/upload.go`). The first
  fix added sanitisation plus a lexical prefix containment check — which a
  pre-existing symlink inside the root defeated: `<root>/uploads -> /tmp/x`
  passed the string check while every write and delete resolved through the
  link. Containment is now enforced by the kernel via `os.Root`
  (openat2-based, no symlink traversal), not by string comparison
- **Filename truncation cut mid-rune.** The 255-byte cap from the upload fix
  sliced names by byte: a multi-byte name ending at the cap gained an invalid
  trailing byte that went to disk and back in `FileInfo.FileName`, and a cut
  across the extension boundary turned `x.png` into an extension-less blob.
  Truncation now walks runes and preserves the extension (`middleware/upload.go`)
- **CORS wildcard subdomains matched too much.** `*.example.com` accepted
  `notexample.com` and `evil-example.com`, because the suffix was not required
  to be preceded by a dot (`middleware/cors.go`)
- **Rate limiting could be bypassed.** `DefaultKeyFunc` keyed buckets on
  `X-Forwarded-For`/`X-Real-IP` — first the whole header, then just its first
  element — but those headers are attacker-controlled either way: a client
  rotating a forged value got a fresh bucket per request, and `MaxKeys` only
  caps memory, not the bypass. The default key is now the TCP remote address;
  servers behind a reverse proxy declare its ranges in the new
  `TrustedProxies` config, which resolves the client through the chain with
  the same right-to-left walk as the IP filter (`middleware/ratelimit.go`)
- **A rejected JWT produced a successful response.** `Router.ServeHTTP` ignored
  the return value of `ExceptionFilter.Capture`, whose contract says false lets
  the exception propagate, so a declining filter left the client with 200 and an
  empty body (`router/router.go`)
- **Slowloris.** No `ReadHeaderTimeout` was set, so a client dribbling headers
  held connections open indefinitely (`server/http.go`)
- **A recovered gRPC panic was reported as success.** The four interceptors
  recovered without setting an error, leaving the named results at `(nil, nil)`.
  They now return `codes.Internal` and log a stack trace (`grpc/grpc.go`)
- **`gofault new` wrote outside the working directory.** The project name was
  not validated, so `gofault new ../evil` scaffolded elsewhere
  (`cmd/gofault/main.go`)
- **Compression and the cache middleware corrupted error responses.** Both swap
  `ctx.Response` for their own writer and neither restored it, so the router
  handed the exception filter a *Ctx pointing at a buffer nobody reads. A failed
  request came back as 200 with an empty body. Every middleware unit test passed
  while the assembled framework was broken; only an end-to-end test could see it
- Compression also committed the capture's default 200 status line on the error
  path, turning the filter's write into a no-op
- **X-Forwarded-For could still be spoofed through a trusted proxy.** The
  filter trusted the leftmost XFF entry, but proxies append their hop rather
  than replacing the header: a client sending `X-Forwarded-For: 127.0.0.1`
  arrived as `127.0.0.1, <real client>` and walked past an allow-list that
  permits localhost. The chain is now walked right to left and the first
  untrusted hop wins; a malformed hop stops the walk rather than trusting
  across it (`middleware/ipfilter.go`)
- **The cache middleware served one caller's authenticated body to every
  later caller.** The key was Host+Method+Path+Query only, so an
  authenticated response cached for the first caller was replayed to other
  users — and to unauthenticated ones — of the same URL. Authorization and
  Cookie are now hashed (SHA-256) into the key, and a response whose own
  headers say `no-store`/`no-cache`/`private`/`max-age=0`, or that carries
  `Vary`, is never cached (`middleware/cache.go`)
- **CORS emitted the invalid `*` + credentials pair.** With
  `AllowOrigins: ["*"]` and `AllowCredentials: true` the middleware set
  `Access-Control-Allow-Origin: *` together with
  `Access-Control-Allow-Credentials: true` — the Fetch spec forbids the
  combination and every browser rejects it, so the configuration worked for
  nobody. With credentials the wildcard now echoes the request's exact
  origin and the response carries `Vary: Origin` (`middleware/cors.go`)
- **A half-configured gRPC TLS setup served plaintext.** `NewServer` builds
  credentials only when *both* `TLSCert` and `TLSKey` are set; with exactly
  one set, credentials stayed nil and the server listened in cleartext with
  no error and no log — the operator believed the endpoint was encrypted.
  `bind` and `Serve` now reject the half-configured setup (`grpc/grpc.go`)

### Concurrency and correctness

- **Concurrent requests ran each other's middleware.** `ServeHTTP` built the
  chain with `append(r.middleware, ...)`, which writes into the router's backing
  array whenever it has spare capacity, so simultaneous requests clobbered each
  other's per-route middleware (`router/router.go`)
- **A disabled middleware panicked every request.** Every constructor returns
  nil when its config is disabled and the documented usage registers whatever it
  returns, so a nil landed in the chain and `runChain` called it. Nil entries are
  now dropped in `Module.RegisterMiddleware`, `Router.Middleware` and
  `Router.Handle`. This was reachable before v3.4.1
- **Module start order was non-deterministic.** `sortModules` walked a map, so
  Go's randomised iteration decided the order: the same dependency graph
  produced five different orders across forty runs (`module/module.go`)
- **Three data races.** `Router.container`, and `App.server`/`App.booted` which
  `Start` wrote while a signal handler calling `Stop` read them
- **`ioc` raced on singletons.** `entry.inst` is written under the container
  mutex but read without it, so concurrent first-resolves raced on the field and
  the fast path could observe a partially written interface value
- **Nested request-scoped resolutions built duplicate instances.** A
  request-scoped type resolved as a constructor argument received a fresh
  `context.Background()` instead of the request context being resolved, so the
  same request got two different instances of one request-scoped type — and
  singletons resolved during a request captured that request's instance for the
  process lifetime. `callCtor`/`resolveType` now thread the request context
  through nested resolutions; singleton constructors get an explicit background
  context (`ioc/container.go`)
- **`App.Init` sorted modules without the lock, and bootstrap never wired the
  container into the router.** `Init` mutated `a.modules` through `sortModules`
  while `RegisterModules` could append concurrently (race detector confirmed);
  and `bootstrapLocked` built the router without `SetContainer`, so
  `ResolveFromCtx` in a controller always failed even though `Init` had built a
  working container (`module/module.go`)
- **The gRPC module never joined the app lifecycle.** The framework iterates
  `OnBootHooks`/`OnShutdownHooks`; it never discovers methods named
  `OnBoot`/`OnShutdown` on a module. `NewGRPCServerModule` defined those
  methods but registered nothing, so the documented "listens on OnBoot,
  drains on OnShutdown" never happened — registering the module into an App
  started no gRPC server and shutdown drained nothing. The constructor now
  registers the module into both hook lists (`grpc/grpc.go`)
- **The timeout middleware replayed a failed handler's partial output.** On the
  non-timeout branch the buffer was flushed onto the real writer even when the
  handler returned an error, committing a 2xx status line the router's error
  path could not retract — the client got the half-written body instead of the
  exception filter's response (`middleware/timeout.go`)
- **Compression's `Flush` sent an implicit 200 and dropped the response.**
  `compressCapture.Flush` delegated to the real writer without replaying the
  buffered body; in net/http `Flush` commits the response with an implicit 200,
  so the handler's chosen status and the whole body were stranded in the buffer
  and discarded on an already-committed writer. `Flush` now replays the buffer
  first and is incremental (a second `Flush` only writes what came since); once
  a handler has flushed, the middleware tops up the remaining bytes instead of
  gzip-compressing and replaying the whole buffer, which would have sent the
  flushed prefix twice (`middleware/compression.go`)
- **Bind failures hung the process.** `StartWithGracefulShutdown` bound the port
  inside a goroutine, so a port already in use was discarded and the call blocked
  forever waiting for a signal (`server/http.go`)
- **Graceful shutdown could not be retried.** It marked the server stopped
  before attempting the shutdown, so a timeout looked like success on the next
  call
- **An exception filter's write could be silently dropped.** Compression
  committed a status line even when the handler failed, so the filter's error
  write became a no-op and the client saw 2xx for a failed request
- **`Wrap` lost the HTTP status.** `IsHTTPException` used a direct type
  assertion, so an HTTPException wrapped with `fmt.Errorf("%w")` was reported as
  500 instead of its own status (`exception/exception.go`)
- **`InvokeHandler` panicked on a wrong signature.** A bare type assertion on
  the controller action took down request handling
- **HEAD and 405 semantics were wrong.** HEAD 404ed against a GET route, and a
  path that matched under a different method reported 404 instead of 405 with an
  `Allow` header (`router/router.go`)

### Resource leaks
- **The session store grew without bound.** A client that never returns the
  cookie is issued a new session on every request, `GetSession` reported an
  expired session as invalid without removing it, and nothing reaped the
  remainder, so the map grew for as long as the process served traffic. Expired
  sessions are now dropped when read, a reaper runs until Close, and MaxSessions
  (default 10000) caps the map. Found by the soak test, not by reading: heap grew
  405 MiB/min under load and is flat after the fix.
- **`SessionConfig.MaxAge` of 0 expired the session immediately.** The field
  documents 0 as "until browser closes", but `now.Add(0)` produced an expiry of
  "now". A zero MaxAge now leaves Expires unset, which the expiry checks skip.

- **`NewRateLimiter` leaked a goroutine per instance.** The cleanup loop had no
  exit condition and there was no way to stop it: 200 limiters left 200
  goroutines behind. Added an idempotent `Close`
- **The rate limiter's key map grew without bound.** Timed-out buckets were only
  reaped by a 5-minute ticker, so rotating keys outpaced it. Added `MaxKeys`
  (default 10000) with least-recently-used eviction
- **`redis.NewClient` leaked its pool on failure.** The connectivity probe
  returned an error without closing the client. Verified: 50 failed
  constructions now leave the goroutine count unchanged
- **`WebSocketHijackMiddleware` leaked upgraded connections.** After
  `upgrader.Upgrade` succeeds the connection is off net/http's management, and
  the middleware never closed it: a handler that returned an error or panicked
  left the socket open until the peer timed out. `WebSocketMiddleware` already
  closed its connection; the hijack variant now does the same on every path
  (`middleware/websocket.go`)

### Configuration that did nothing

Each of these was documented as working and read by nothing. Go is silent about
it, so `internal/deadfield` now fails the build on both this and the next class.

- `gorm.Config.Silent` never suppressed logging; it now selects `logger.Discard`
- `exception.IncludeStackTrace` never populated a stack; it now does, off by
  default since the stack exposes internal paths
- `versioning.HeaderVersionFormat` was stored and never applied;
  `HeaderConfig` now derives a regex from it as documented
- `static.FollowSymLinks` and `ipfilter.Mode` are covered above under Security
- `InvokeHandler` accepted an HTTP method and path and read neither, so its
  signature implied a dispatch that never happened. Both parameters removed
- `middleware.ValidatorConfig` documented a validator middleware that does not
  exist; nothing read it. Removed

### Other correctness

- **The scaffold did not compile.** `main.go` imported `<name>/controllers` as
  `github.com/gofault/gofault/<name>/controllers`, and the generated controller
  returned `(int, error)` from a method declared `error`. `testapp3` carried the
  same two defects and, being a separate module, was never compiled by
  `go test ./...` (`cmd/gofault/main.go`)
- **`metrics.normalizePath` was a stub**, so every distinct ID became a new
  Prometheus series
- **The request logger built its attributes and discarded them**, producing no
  output at all, and `readRequestBody` returned an empty string
- **A duplicate DI registration silently replaced the first**, so a later typo
  looked like it had taken effect. Now returns `ErrDuplicateRegistration`
- **Configuration was never validated.** An unknown key such as `sever.port` was
  ignored, leaving the server on its default port. `Load` now rejects unknown
  fields and validates ports, drivers and log settings
- **`GetDSN` always emitted the MySQL form**, which PostgreSQL and SQLite drivers
  cannot parse
- **The cache capture broke streaming and hijacking, and buffered without
  bound.** `responseCapture` embedded `http.ResponseWriter` and defined only
  `Write`/`WriteHeader`/`Header`, so it did not satisfy `http.Flusher` or
  `http.Hijacker` — embedding does not extend the method set — and handlers
  that type-assert those interfaces failed outright under the cache
  middleware. It also appended every byte with no cap, so one large download
  pinned its whole body in a cache entry. The capture now implements both
  interfaces (a Flush or Hijack streams the response and opts it out of the
  cache) and caps buffering at `MaxBodyBytes` (default 256 KiB), streaming
  anything larger (`middleware/cache.go`)
- **Compression broke WebSocket upgrades.** `compressCapture` likewise did not
  satisfy `http.Hijacker`, and the upgrader type-asserts the writer to it —
  with compression enabled and a client sending `Accept-Encoding: gzip` (every
  browser does), the handshake failed outright. Upgrade requests now bypass
  the capture entirely: a 101 handshake is followed by raw protocol bytes on
  the same connection and must never be buffered (`middleware/compression.go`)
- **DSN credentials were interpolated without escaping.** A password containing
  `@` or `:` — the usual shape of a generated secret — moved the credential
  boundary in both formats, and a database name containing `?` or `/` broke
  MySQL parsing outright. The mysql branch now percent-encodes the database
  name (the driver decodes `%XX` there), rejects a username containing `:`
  because the driver splits the userinfo at the first colon and does not
  decode escapes, and leaves `@ : # &` in the password alone — verified
  against go-sql-driver v1.8.1's actual parser. The postgres branch builds the
  URL through `net/url` so the userinfo and path are RFC 3986 encoded
  (`config/config.go`)
- **Compression was O(n²)** for large responses, accumulating into a `[]byte` by
  repeated append
- **The OpenAPI document never listed a route.** `AddPath` existed but was never
  called, so the published spec had no paths
- **A panicking health checker took down the health endpoint**; it is now
  reported as down
- **`gofault version` reported a hardcoded `v1.0.0`**

### Added
- `middleware.CacheBackend`, so a cache need not live in the process.
  `*InMemoryCache` satisfies it, so existing callers are unaffected, and
  `redis.HTTPBackend` makes the integration the old comment promised
- `internal/deadfield`, an AST-based guard that fails the build on an exported
  field nothing reads and on a function parameter never used. Exempt entries
  need a reason and a second test fails when an exemption goes stale
- `server.HTTP.Server()` for advanced configuration
- `OpenAPIHandler`, `RegisterOpenAPI`, `HealthCheckHandler` and
  `RegisterHealthCheckEndpoint`, which separate the handler and middleware roles
  the old names conflated
- A `submodules` CI job that builds and vets `testapp3`, invisible to a
  root-level `go test ./...`
- End-to-end tests (`-tags e2e`) driving the assembled framework: a real server,
  the router, a full middleware chain and real concurrent clients, plus fault
  injection for aborted and slow requests
- Integration tests (`-tags integration`) against real MySQL and PostgreSQL, so
  the generated DSN is parsed by an actual driver rather than compared as a
  string. Both suites run in CI and skip locally when the servers are absent

### Changed

- JWT: query-string tokens are opt-in via `AllowQueryToken`, and tokens without
  an `exp` claim are rejected unless `RequireExpiry` is disabled. Both defaults
  are safer; `LegacyJWTConfig` restores the old behaviour for a deliberate
  migration
- `JWTAuth` returns a middleware that rejects requests with a diagnostic for an
  unusable config, instead of nil. `MustJWTAuth` is available for callers who
  prefer to fail fast at setup
- `CacheMiddleware` takes a `CacheBackend` rather than a concrete type
- `server.HTTP.Start` returns nil rather than `http.ErrServerClosed` after a
  graceful shutdown
- `InvokeHandler` takes only the handler name
- CLI: the scaffold pins `gofault v0.0.0`, resolved by the local `replace`,
  instead of the non-existent `v1.0.0`, and targets the current Go version

### Removed

- `middleware.ValidatorConfig`, which had no consumer and no effect.
  `ValidateRequest(ctx, rules)` remains the supported API; route-scoped
  validation with struct binding is not implemented

### Architecture risks addressed

- **Multi-replica rate limiting.** `middleware.RateLimiter` counts per process,
  so with N replicas the effective limit is N times the configured one and a
  client that reaches two replicas gets it twice. That is now documented with a
  comparison table, the fixed-window boundary burst and the fail-open behaviour
  spelled out, and guidance on when a per-process limiter is still the right
  choice. `redis.DistributedLimiter` shares the count in Redis.
- **Key map capacity.** `MaxKeys` bounds how many buckets are held. The README
  now gives the memory figure per key and what happens at the cap, since the
  background reaper alone cannot keep up with a client rotating keys.
- **Redis outage recovery.** Verified against a real container that is stopped
  and restarted mid-test: the limiter fails open during the outage and resumes
  enforcing afterwards rather than staying permanently disabled, the cache
  degrades to errors and recovers, and a long-lived client reconnects on its own.
- **Connection pool exhaustion.** The contract is now documented: an exhausted
  pool blocks until the caller's context expires and reports
  `context deadline exceeded`, so a busy pool is distinguishable from an
  unreachable database. Verified against real MySQL, along with recovery on
  release, a burst of 40 concurrent callers against a 4-connection pool, and that
  an abandoned request returns its connection instead of wedging the pool.

### Added
- A soak suite (`-tags soak`) driving sustained mixed traffic through the
  assembled framework: a real server, request IDs, recovery, compression, rate
  limiting and sessions. It samples heap and goroutine counts under load, fits a
  line to the second half, and fails on a materially positive slope. Locally
  `GOFAULT_SOAK_DURATION=25s`, in CI three minutes on the main branch.
- The soak also asserts the session cap does not strangle throughput. Enforcing
  the cap with an O(n) scan on every insert cut the request rate by about eight
  times, which a growth-only check cannot see.

### Test results

Statement coverage 50.0% -> 87.1%. `controller` 31.6% -> 100%, `core` 58.3%
-> 100%, `exception` 77.1% -> 98.2%, `redis` 78.8% -> 88.5%, `grpc` 81.6%
-> 91.3%, `gorm` 74.1% -> 83.3%, `server` 56.0% -> 90.0%.

gRPC additionally gained end-to-end tests over bufconn with a real client;
RegisterService, the entry point for using the server at all, had no coverage
and the test named after it never called it.

Every fix has a test that fails when the fix is reverted. Notable additions:
randomised coverage of the topological sort over 200 generated dependency
graphs, a concurrency test driving 10000 requests through a middleware stack
while watching the goroutine count, and a test that compiles a freshly
scaffolded project.
