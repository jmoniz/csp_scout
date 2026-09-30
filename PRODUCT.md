# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Primary: a single developer or engineer, working on their own machine, who owns or is responsible for a web application that is moving toward a Content Security Policy. They have real traffic or a realistic staging environment and need to know what would actually break before enforcement, so they run the target site in `Content-Security-Policy-Report-Only` mode and let the browser report violations.

Job to be done: turn an unbounded stream of real CSP violation reports into a small set of allow/deny decisions, then into a policy they can paste into their server, CDN, or HTML.

Secondary: a security engineer auditing a site they do not control, injecting a report-only header through a browser extension for the duration of a session.

## Product Purpose

Capture CSP violations reported by browsers, normalize them into directive-and-origin groups a human can actually triage, and compile the resulting decisions into a production-ready policy.

The product exists to make enforcement safe: report-only first, evidence second, policy last. Success means a developer moves from "CSP might break my app" to a deployed policy where every allowed source traces back to an observed violation the developer approved.

## Positioning

Three claims a neighboring tool could not honestly copy:

1. **Nothing leaves the machine.** Completely local and offline by default. Violation data, target origins, and decisions never reach a third-party service, and there is no telemetry to turn off.
2. **One flow, one tool.** Capture, per-directive/per-origin triage, and multi-format policy export happen in the same place, with the policy updating live as decisions are made. No export/import shuttle between a collector, a spreadsheet, and a policy evaluator.
3. **It deduces, it does not dictate.** The tool derives the rule a developer would otherwise work out by hand — `https://api.example.com` to `https://*.example.com`, and detection of origins that are really `'self'` — and offers it for approval. It never silently widens a policy.

## Operating Context

- Runs locally: Go backend on `:8080`, Vite dev server on `:5173`, or both containers via `docker compose` on `:3000` / `:8080`. SQLite database on the local filesystem.
- The target site is instrumented from the browser, not its server: a browser extension (Lazy Header Editor) injects the `Content-Security-Policy-Report-Only` response header and a per-session Report-URI, filtered by URL. Reloading the site produces real violations.
- Ingests legacy W3C `application/csp-report`, Reporting API v1 `application/reports+json`, and plain JSON. Complies with W3C CSP Level 2 and 3 output rules (keywords quoted, host sources and schemes unquoted).
- A session targets exactly one web origin and carries its own policy settings; multiple sessions coexist for multiple sites or environments.
- Export formats: HTTP header, HTML `<meta>`, Nginx `add_header`, Apache `Header set`, and raw directives, with an enforce versus report-only toggle.

## Capabilities and Constraints

Confirmed:

- Session management (create, list, delete with cascade), per-session Report-URI and header snippet, plus a copy-ready browser-extension setup guide.
- Ingestion with CORS preflight, origin normalization, wildcard deduction, and `'self'` detection against the session target origin.
- Violation triage by directive tab, with per-origin approve / approve-wildcard / mark `'self'` / reject / ignore, bulk approval, a pending-queue "inbox zero" workflow, and an evidence drawer showing raw occurrences, source scripts, and line and column numbers.
- Live capture mode with polling while reports arrive, and a live multi-format policy preview.
- Pure-Go embedded SQLite (`modernc.org/sqlite`) with WAL mode; no CGO, no external database, no external services.
- English is the language of the UI, documentation, and code comments (explicit project constraint).

Constraints and explicitly undecided facts:

- **Single-user, local, no authentication.** There is no user model, no session isolation beyond the session entity, and CORS is currently wide open. Correct for the current phase: a phase constraint, not a permanent one, and the first thing that must change if the tool is ever hosted for a team.
- **Distribution is undecided.** Internal, open source, or commercial has not been chosen; there is no license file and no git remote.
- **Accessibility standard is not established.** No product-specific requirement has been confirmed; do not assume WCAG conformance as a product promise.

## Brand Commitments

- **Name: CSP Scout.** Confirmed as canonical. The strings currently in the repository are legacy and are not authorities: "CSP Shield" in the UI chrome and `<title>`, and "CSP Report Collector & Policy Generator" in README and footer. The rename has not yet been applied to the code.
- There is no logo or wordmark. The only mark in the product is an inline generic shield SVG icon, which is not a brand asset and must not be treated as one.
- **Voice: precise and technical, aimed at an engineer.** Terminology already settled in the product: *session* (one target origin under analysis), *violation group* (aggregated by directive and origin), *triage status*, *policy settings*. Does not cheerlead, does not use marketing superlatives.

## Evidence on Hand

- Implementation history: all four original plan phases (Go backend, SvelteKit frontend, Docker packaging, E2E testing) are marked complete.
- `README.md` — feature summary, local and Docker start-up, browser instrumentation guide.
- `backend/` — Go API, parser, generator, and their unit and integration tests.
- `scripts/test_e2e.mjs` — end-to-end verification against a running backend.
- Frontend Vitest/browser assets: none. Frontend verification is build-and-inspect.

Absences future work must not fill in: no license, no git remote, no customers, no testimonials, no press, no benchmarks or performance numbers, no logo or brand asset beyond an inline generic shield SVG icon. Do not fabricate any of these.

## Product Principles

1. **Local and offline by default.** Violation data stays on the operator's machine. Any future feature that ships data off-machine is a break of the product's core claim, not an enhancement.
2. **Evidence before enforcement.** Every rule that reaches the generated policy traces back to an observed violation the human approved. The tool's default posture is report-only.
3. **The tool proposes, the human decides.** Deduction of wildcards and `'self'` exists to remove manual work, never to widen a policy without an explicit approval.
4. **One place for the whole loop.** Capture, triage, and export stay in a single continuous workflow so the policy can update live as decisions are made.
5. **Every decision is inspectable.** Raw occurrences, source files, and line numbers stay reachable from the triage view; a triage action without visible evidence behind it is a bug.
