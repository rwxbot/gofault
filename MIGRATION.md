# Migrating to v3.4.1

v3.4.1 is a defect-fix release: 78 defects across 35 commits (see
`CHANGELOG.md`). Most fixes are invisible — the framework starts doing what the
documentation always said it did. A minority change observable behaviour on
purpose, because the old behaviour was a security hole or a silent failure.
This document lists every change an application can notice, split by what you
must do about it.

If the upgrade compiles, you are **not** automatically done: several changes
are runtime-behaviour changes that a clean build will not surface. Work
through "Behaviour changes" below.

## Breaking changes (the build will tell you)

These produce compile errors, so they are impossible to miss.

### `InvokeHandler` lost two parameters

```go
// before
controller.InvokeHandler(ctrl, method, path, ctx)
// after
controller.InvokeHandler(ctrl, handlerName, ctx)
```

The HTTP method and path were never read — routing has already happened by the
time a handler runs. The old signature implied a dispatch that never occurred.

### `middleware.ValidatorConfig` is removed

It documented a validator middleware (`BindTarget`, `SkipMissing`) that no code
ever read; setting it did nothing. The supported API is unchanged:

```go
if err := middleware.ValidateRequest(ctx, middleware.Rules{
    "name": {&middleware.Required{}},
}); err != nil { ... }
```

Route-scoped validation with struct binding is **not** implemented; if you
relied on `ValidatorConfig` compiling, you were relying on nothing.

### `CacheMiddleware` takes an interface, not a concrete type

```go
// before
middleware.CacheMiddleware(cache *InMemoryCache, cfg CacheConfig)
// after
middleware.CacheMiddleware(backend CacheBackend, cfg CacheConfig)
```

`*InMemoryCache` satisfies `CacheBackend` structurally, so existing call sites
keep compiling. The point of the change is that a Redis-backed backend can now
be shared across instances:

```go
cache := redis.NewCache(rdb, "myapp:httpcache:", 5*time.Minute)
backend := redis.NewHTTPBackend(cache, ctx) // ctx bounds every operation
middleware.CacheMiddleware(backend, cfg)
```

## Behaviour changes (the build will NOT tell you)

### JWT defaults are stricter

Two defaults flipped, both because the old behaviour leaked credentials:

- **Query-string tokens are now rejected.** A token in a URL lands in access
  logs, `Referer` headers and browser history. Set
  `JWTConfig.AllowQueryToken = true` to accept them again.
- **Tokens without an `exp` claim are now rejected.** A token without an
  expiry never stops working. Set `RequireExpiry = false` to accept them
  again. Note this checks only that `exp` is present — validating lifetime
  separately is your application's job.

To restore the pre-v3.4.1 behaviour wholesale, use the escape hatch:

```go
middleware.JWTAuth(middleware.LegacyJWTConfig(secret))
// equivalent to: AllowQueryToken=true, RequireExpiry=false
```

Use it as a migration bridge, not a destination — both settings weaken security.

`JWTAuth` also no longer returns `nil` for an unusable config (empty secret,
non-HS256 algorithm). It returns a middleware that rejects every request with
a logged diagnostic, so a misconfiguration cannot silently leave routes
unprotected. If you prefer to fail at setup, use `MustJWTAuth(cfg)`, which
panics like the old `JWTMiddleware` helper did.

### Rate limiting keys on the TCP connection, not on headers

`DefaultKeyFunc` used to key buckets on `X-Forwarded-For` / `X-Real-IP`.
Those headers are attacker-controlled: a client rotating a forged value got a
fresh bucket per request, defeating the limit entirely. The default key is now
`RemoteAddr`.

If your server sits behind a reverse proxy, declare the proxy ranges or the
real client IP disappears into the proxy's address:

```go
cfg := middleware.RateLimiterConfig{
    TrustedProxies: []string{"10.0.0.0/8", "127.0.0.1"},
}
```

With trusted proxies configured, the XFF chain is walked right to left and the
first untrusted hop wins — the same resolution as the IP filter (see below).
`MaxKeys` (default 10000) bounds the key map with LRU eviction; past the cap
the oldest buckets are dropped. `NewRateLimiter` used to leak its cleanup
goroutine; call `limiter.Close()` (idempotent) when you are done with a
limiter instance.

Note the limiter is still per-process: with N replicas the effective limit is
N times the configured one. For a shared count use `redis.DistributedLimiter`.

### The IP filter resolves XFF right to left

The filter used to trust the leftmost `X-Forwarded-For` entry. Proxies append
their hop rather than replacing the header, so a client sending
`X-Forwarded-For: 127.0.0.1` arrived as `127.0.0.1, <real client>` and walked
past an allow-list that permits localhost. The chain is now walked right to
left and the first untrusted hop wins; a malformed hop stops the walk rather
than trusting across it. Allow-lists that were "working" by accidentally
trusting a spoofed leftmost entry may now deny the real client — configure
`TrustedProxies`/trusted ranges so the walk starts from the right hop.

### Static file serving is contained

- The containment check used `HasPrefix(absPath, absDir)`, so root `/srv/www`
  also served `/srv/www-secret/...`. Containment is now exact per path segment.
- `FollowSymLinks` was documented but never read; it is now honoured. With it
  false (the default), a symlink inside the root pointing outside it is no
  longer served.
- Range serving and directory listings now propagate write errors through the
  middleware chain instead of discarding them. A half-written body (client
  gone, broken pipe mid-range) surfaces as an error rather than looking like
  success in logs and metrics.

### Uploads are kernel-contained

The upload storage root is enforced with `os.Root` (openat2, no symlink
traversal) rather than string comparison, so a pre-existing symlink inside the
root (`<root>/uploads -> /tmp/x`) can no longer carry writes outside it.
Filename truncation at 255 bytes is rune-safe and preserves the extension.
Filenames that used to land outside the root now fail; names cut mid-rune by
the old byte cap now truncate cleanly.

### CORS wildcard + credentials

`AllowOrigins: ["*"]` together with `AllowCredentials: true` used to emit the
`Access-Control-Allow-Origin: *` + `Access-Control-Allow-Credentials: true`
pair, which the Fetch spec forbids and every browser rejects — the
configuration worked for nobody. With credentials, the wildcard now echoes the
request's exact origin and adds `Vary: Origin`. Separately, `*.example.com` no
longer matches `notexample.com` / `evil-example.com`: the suffix must be
preceded by a dot.

### Cache middleware: key, policy, and buffering

- **The cache key now includes the caller.** Authorization and Cookie headers
  are hashed (SHA-256) into the key. Previously an authenticated response
  cached for the first caller was replayed to every later caller of the same
  URL, including unauthenticated ones. If you were relying on cross-user
  sharing of a cached URL, that was the vulnerability; cache public responses
  on routes that do not run through authenticated handlers, or vary upstream.
- **Cache-control is honoured.** A response whose own headers say
  `no-store`/`no-cache`/`private`/`max-age=0`, or that carries `Vary`, is
  never cached.
- **Buffering is capped.** `CacheConfig.MaxBodyBytes` (default 256 KiB):
  larger responses stream through and opt out of the cache.
- **Streaming and hijacking work.** The capture now implements `http.Flusher`
  and `http.Hijacker`; a handler that type-asserts them no longer fails under
  the cache middleware, and a Flush or Hijack opts that response out of
  caching.

### Compression: WebSocket upgrades and Flush

- Upgrade requests bypass the compression capture entirely. Previously, with
  compression enabled and `Accept-Encoding: gzip` (every browser sends it),
  the WebSocket handshake failed outright.
- `Flush` now replays the buffered body before delegating, and is
  incremental. Previously it committed an implicit 200 and stranded the real
  status and body in the buffer.

### Timeout middleware no longer replays failed output

On the non-timeout branch the buffered body used to be flushed even when the
handler returned an error, committing a 2xx status line the router's error
path could not retract — the client got a half-written body instead of the
exception filter's response. Failed handlers now produce the filter's
response.

### gRPC

- **Half-configured TLS is refused.** Exactly one of `TLSCert`/`TLSKey` used
  to fall through to plaintext with no error and no log. `Start` and `Serve`
  now return an error naming the problem. If you were serving plaintext
  accidentally, set `Insecure: true` explicitly.
- **The module lifecycle actually runs.** `NewGRPCServerModule` used to define
  `OnBoot`/`OnShutdown` without registering them, so adding the module to an
  App started no server. The constructor now registers both hooks. `OnBoot`
  uses `StartInBackground`, which binds synchronously — bind failures return
  from the call instead of hanging `App.Start`.
- **Recovered panics become errors.** The four interceptors used to recover
  and return `(nil, nil)`, reporting the call as a success with a nil
  response. They now return `codes.Internal` and log a stack trace.
- **`Start` binds `:PORT`, not `tcp:PORT`.** The old address construction made
  every default-config `Start()` fail. If you worked around it by passing an
  explicit `Network`/`Port`, re-check your config.

### DI: duplicate registrations are errors

Registering two constructors for the same key used to silently replace the
first, so a later typo looked like it had taken effect. The second
registration now returns `ioc.ErrDuplicateRegistration`. Audit startup for
duplicate keys — the framework's own module wiring deliberately ignores this
error for controller bindings (all controllers share one key; routes are wired
directly), but application code should not.

### Configuration files are validated

`config.Load` now rejects unknown fields (`config.ErrUnknownField`) and
validates ports, drivers and log settings. A typo like `sever.port` used to be
ignored silently, leaving the server on its default port. If your config
carries keys the framework never had, Load now fails — remove them.

### Sessions

- The store no longer grows without bound: expired sessions are dropped when
  read, a reaper runs until `Close`, and `MaxSessions` (default 10000) caps
  the map. Past the cap, new sessions are refused.
- `SessionConfig.MaxAge = 0` now means "until browser closes" as documented;
  previously it expired the session immediately. If you were compensating for
  the bug by setting a large MaxAge, 0 is now the correct value.

### HTTP semantics

- **405, not 404.** A path that matches under a different method now returns
  405 with an `Allow` header. Clients (and tests) that asserted 404 for wrong
  methods need updating.
- **HEAD works against GET routes** instead of 404ing.
- **`StartWithGracefulShutdown` reports bind failures.** A port already in
  use used to be discarded inside a goroutine while the call blocked forever
  waiting for a signal; the bind error now returns immediately.
- **`server.HTTP.Start` returns `nil` after a graceful shutdown** instead of
  `http.ErrServerClosed`. Code that branched on `errors.Is(err,
  http.ErrServerClosed)` to detect a normal shutdown should treat a nil
  return as normal.
- **Exception filters are respected.** `Router.ServeHTTP` used to ignore the
  return value of `ExceptionFilter.Capture`; a filter declining to handle an
  exception left the client with 200 and an empty body. Returning false now
  propagates the exception.
- **Disabled middlewares are safe to register.** Every constructor returns nil
  when its config is disabled, and the documented usage registers whatever it
  returns — a nil entry used to panic every request. Nil entries are now
  dropped by `Module.RegisterMiddleware`, `Router.Middleware` and
  `Router.Handle`.

### DSN generation

`GetDSN` used to always emit the MySQL form (unparseable by postgres/sqlite
drivers) and interpolated credentials without escaping — a password containing
`@` or `:` moved the credential boundary. The mysql branch now percent-encodes
the database name and rejects a username containing `:`; the postgres branch
builds the URL through `net/url` (RFC 3986). DSNs you generated and stored
elsewhere may need regenerating.

### Configuration that did nothing now does something

These fields were documented, read by nothing, and now take effect. If you set
them expecting no effect (or worked around their inertness), review them:

| Field | Now does |
|---|---|
| `gorm.Config.Silent` | Suppresses all GORM logs (`logger.Discard`) |
| `exception.IncludeStackTrace` | Actually populates stack traces; **off by default**, since stacks expose internal paths |
| `versioning.HeaderVersionFormat` | Derives the header regex as documented |
| `static.FollowSymLinks`, `ipfilter.Mode` | See the security sections above |
| `middleware/cache.CacheConfig.MaxBodyBytes` | Caps buffered bodies (default 256 KiB) |

An AST guard (`internal/deadfield`) now fails the build on any exported field
nothing reads, so this class cannot come back.

### CLI scaffold

`gofault new` pins `github.com/gofault/gofault v0.0.0` resolved by the local
`replace` instead of the non-existent `v1.0.0`, and the generated `go.mod`
targets the current Go version. The scaffolded project compiles out of the box
(a test compiles it). Project names are validated: `gofault new ../evil` no
longer scaffolds outside the working directory.

## Opt-in additions

Nothing here changes existing behaviour; adopt as needed.

- `middleware.CacheBackend` + `redis.NewHTTPBackend` — shared HTTP cache
  across instances (see Breaking changes above).
- `OpenAPIHandler` / `RegisterOpenAPI`, `HealthCheckHandler` /
  `RegisterHealthCheckEndpoint` — separate the handler and middleware roles
  the old names conflated. The OpenAPI document now actually lists routes
  (`AddPath` was never called before).
- `server.HTTP.Server()` — access to the underlying `http.Server` for advanced
  configuration (e.g. tuning the new `ReadHeaderTimeout`, which is set to
  close Slowloris-style header dribbling).
- `redis.DistributedLimiter` — cross-replica rate limiting (the in-process
  limiter's N-replica caveat is now documented).
- `middleware.WebSocketHijackMiddleware` no longer leaks upgraded connections
  on handler error/panic paths.

## Suggested upgrade order

1. Update the two compile-breaking call sites (`InvokeHandler`, any
   `ValidatorConfig` reference).
2. Run the app's test suite — several behaviour changes (405, cache key,
   timeout replay) surface as assertion failures.
3. Start the app and watch startup logs: config validation errors,
   `ErrDuplicateRegistration`, gRPC TLS refusals.
4. If behind a reverse proxy, set `TrustedProxies` on the rate limiter and
   IP filter before trusting their keys/decisions again.
5. Only if necessary, bridge JWT with `LegacyJWTConfig` and plan its removal.
