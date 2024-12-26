# Architecture

gofault is a small framework with one rule: **dependencies point at `core`**.
`core` is the only package with no framework imports, so nothing can create a
cycle and every other package stays swappable.
