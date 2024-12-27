# Architecture

gofault is a small framework with one rule: **dependencies point at `core`**.
`core` is the only package with no framework imports, so nothing can create a
cycle and every other package stays swappable.

## Layers

```
core                       contracts: Ctx, Module, Route, Provider,
                           Controller, Handler, MiddlewareFunc, hooks
├── ioc                    container: singleton / request / transient
├── provider               injectable values and request-scoped accessors
├── router                 method + path dispatch, middleware chain,
│                          exception filters
├── server                 net/http listener, graceful shutdown
├── logger, config         infrastructure helpers
├── exception              HTTP errors and the filter that renders them
├── versioning             version negotiation
├── middleware             cross-cutting handlers built on core only
├── controller             prefix routing and response helpers
├── module                 composition root
├── gorm, redis, grpc      optional integrations
└── <your app>             registers modules into module.App
```

The measured import graph (each entry lists the framework packages it imports)
is deliberately shallow:

| package | imports |
|---|---|
| `core`, `ioc`, `server` | — |
| `config`, `provider` | `ioc` (`provider` only) |
| `controller`, `exception`, `logger` | `core` |
| `versioning` | `core` |
| `middleware` | `core`, `exception` |
| `router` | `core`, `exception`, `ioc` |
| `gorm`, `grpc`, `redis` | `core` |
| `module` | `core`, `controller`, `ioc`, `provider`, `router`, `server` |

Nothing imports `middleware`, and nothing imports the integrations. That is why
you can delete `gorm`, `redis` and `grpc` from a build without touching the core.

## Request lifecycle

1. `server.HTTP` accepts the connection and builds a `*core.Ctx` with
   `core.NewCtx`, which initialises `Params` and `Locals`.
2. `router.Router.ServeHTTP` matches method and path, extracts `:param`
   segments into `ctx.Params`, and begins a request scope in the container.
3. The router's middleware chain runs outermost first, then the module
   middleware attached to the matched route.
4. `controller.InvokeHandler` reflects the handler name from the route onto the
   controller and calls it.
5. If a handler returns an error, the router hands it to the exception filter
   chain; `exception.NewHTTPExceptionFilter` recognises `HTTPException` values
   and writes the matching status code and JSON body.
6. The request scope ends, releasing request-scoped instances.
