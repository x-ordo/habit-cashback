# Claude Code / speckit Prompt — Apply v0.7 to your repo

You are working inside the target repository (existing codebase). Apply the changes from this kit **without breaking existing functionality**.

## Goal
Implement Apps-in-Toss flow end-to-end:
- appLogin → auth/exchange (server generates JWT)
- make-payment → checkoutPayment → execute-payment
- promotion: get-key → execute-promotion → execution-result polling (worker)
- messenger: send-message
- mTLS required for all Partner API calls

## Tasks (do in order)
1) Create/Update Go backend module
- Copy `backend/` from this kit into the target repo.
- Ensure `cmd/api` runs and exposes `/health`.
- Add `.env.example` and update README quickstart.

2) DB
- Apply `db/migrations/001_init.sql` (Postgres).
- Add migration runner (optional): `make migrate` or `goose`.

3) Security & Idempotency
- Enforce `Authorization: Bearer <JWT>` for all non-auth endpoints.
- Persist idempotency records (scope + key → response_json).
- For make-payment and execute-promotion, require an idempotency key; if absent generate UUID.

4) Worker finalize
- In `cmd/worker`, implement payout polling:
  - query payout where status in ('REQUESTED','PENDING')
  - call execution-result with the same user's `X-Toss-User-Key`
  - update payout status to SUCCESS/FAIL/PENDING

5) Front integration
- Wire up `appLogin()` and `checkoutPayment({ payToken })` on the client.
- Client should never receive Toss accessToken/refreshToken.
- Client only receives internal JWT.

## Acceptance tests
- Login: `POST /v1/auth/exchange` returns `{ token, userKey }`
- Payment: create returns payToken; checkout opens; execute succeeds; DB has payment rows.
- Promotion: issue returns promotionKey; worker eventually marks SUCCESS.
- Messenger: send returns counts.

## Don’t
- Do not use iframe-based UIs.
- Do not store refreshToken/accessToken in the browser.
- Do not call Toss Partner API without mTLS.
