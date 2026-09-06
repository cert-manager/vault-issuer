# Conformance suites (disabled)

These files register the Vault issuer into cert-manager's shared conformance
suites. They cannot compile here: the harness they need lives in
`github.com/cert-manager/cert-manager/e2e-tests`, whose module path does not
match its directory, so no proxy can serve it.

The directory is named with a leading underscore so the Go tool ignores it. The
code is kept for reference until the harness is importable; see the tracking
issue in the repository README.
