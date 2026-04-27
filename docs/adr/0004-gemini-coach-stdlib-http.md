# ADR-0004: Gemini Coach via stdlib net/http

Date: 2026-04-27
Status: Accepted

---

## Context

The AI coach feature sends a completed game's PGN to Google Gemini and returns
structured feedback (`Analysis` with a summary and up to six `Mistake` objects).

Google provides two integration paths:

**Option A — Official Go SDK** (`google.golang.org/genai`)
Provides type-safe method calls, automatic retries, and managed client
configuration.

**Option B — Direct REST calls via stdlib `net/http`**
Call the `generateContent` endpoint directly. Request and response structs are
defined locally; JSON is marshalled/unmarshalled with `encoding/json`.

---

## Decision

Use Option B: direct REST calls with `net/http`.

The Gemini REST API for `generateContent` has a single relevant endpoint and a
simple request envelope. The local request/response types (`geminiRequest`,
`geminiResponse`, `generationConfig`) total about 25 lines. The call itself is
standard: build request, POST, read body, unmarshal.

The key Gemini feature used is `responseMimeType: "application/json"`, which
instructs the model to emit only valid JSON. This eliminates the need to parse
prose or strip markdown fences in the common case (a defensive `stripCodeFence`
helper handles the rare mismatch).

---

## Consequences

**Positive**
- No SDK dependency in `go.mod`; one fewer transitive dependency tree.
- The implementation is easy to test with an `httptest.Server` stub —
  the `withBaseURL` method on `Client` redirects calls to the test server
  without any mocking framework.
- The prompt and generation config are fully visible in the source file; no
  SDK abstraction hides them.
- `responseMimeType: "application/json"` with `temperature: 0.4` produces
  reliably parseable output in practice.

**Negative / trade-offs**
- If Google changes the REST contract (field names, endpoint path, auth
  mechanism) the local structs must be updated manually. The SDK would absorb
  such changes automatically.
- No automatic retry or back-off on transient 5xx errors; the HTTP client
  respects only the 30-second timeout. A production system should add retry
  logic.
- The `normaliseSeverities` function is a defensive measure against the model
  hallucinating severity values outside the three allowed strings. This guard
  would be unnecessary if a strongly typed SDK enum were used.
