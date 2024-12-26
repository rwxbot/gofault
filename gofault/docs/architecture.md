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