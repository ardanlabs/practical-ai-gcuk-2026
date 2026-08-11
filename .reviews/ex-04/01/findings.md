# Service DiffGuard Report

Scope: branch `ex-04` vs `main` (`git diff main...HEAD`) — 36 files, +2255/-9. Go service scaffolding, user CRUD, goose migration, curl integration harness, makefile.
Overall Gate: **G4 — Stop**

No issue tracker ticket exists for this work. Acceptance criteria were taken from `prompts.txt` (the accumulated requirement log) plus the six commit messages on the branch, and are graded as such.

## Run matrix

| Lens               | Result | Notes                                                                        |
|--------------------|--------|------------------------------------------------------------------------------|
| Security Sentinel  | **G4** | All five user CRUD routes are unauthenticated; internal error text reaches clients. |
| Spec Cartographer  | **G3** | Every prompts.txt criterion implemented; the `test-curl` gate cannot fail on preflight. |
| Service Steward    | G2     | Builds/vets/gofmt clean. `make tidy` vendors a non-vendored repo; unbounded list query. |
| Error Tripwire     | **G4** | Reproduced: integration harness exits 0 after a hard failure.                 |
| Doc Drift Check    | **G3** | `jqv`'s documented contract is false, and a test relies on the documented behavior. |
| Harness Map        | **G4** | Zero Go tests in 16 packages; `make test` is vacuously green.                 |
| Boundary Keeper    | **G3** | `userbus.User.ID` is a primitive `int64` — MUST-level layering violation.     |
| Clone Hunter       | G2     | Both headline duplication candidates refuted; column list repeated 4×.        |
| Straight-Line Pass | G2     | Four exported functions with zero callers; two unreachable branches.          |

`go build ./...`, `go vet ./...` and `gofmt` are clean. Nothing here is a compile break.

## Acceptance-criteria coverage

Criteria derived from `prompts.txt`; there is no ticket.

| Criterion | Verdict | Sites that must implement it | Proof / gap |
|-----------|---------|------------------------------|-------------|
| Three routes bound (healthcheck, Hello, List Users) | **Met** | `app/sdk/mux/mux.go`, one `route.go` per domain | `mux.go:24-26`; `checkapp/route.go:19`, `helloapp/route.go:19`, `userapp/route.go:21-25` |
| Healthcheck replies 200 to GET | **Met** (narrow reading) | `checkapp/route.go`, `checkapp/checkapp.go` | `route.go:19`, `checkapp.go:16-18`; `checkapp_test.sh:15-19`. See weakness S1 |
| Hello GET route, `{USER}` from the path | **Met** | `helloapp/route.go`, `helloapp/helloapp.go` | `route.go:19` (`GET /hello/{user}`), `helloapp.go:20,25`; `helloapp_test.sh:15-36` |
| List Users from postgres via sqlx, rendered as JSON | **Partial** | `userdb.QueryAll`, `userbus.QueryAll`, `userapp.queryAll`, `sqldb.Open`, one assertion | Implemented: `userdb.go:102-117` (`sqlx.SelectContext`), `userbus.go:88-95`, `userapp.go:86-93`, `sqldb.go:45`. Gap: payload-shape assertions self-skip on an empty DB — `userapp_test.sh:26,46-49` |
| Schema moved to `zarf/schemas/db` as a goose migration | **Met** | `zarf/schemas/db/*`, `business/sdk/migrate`, `api/tooling/admin`, makefile | `00001_create_users_table.sql:1,13` (`+goose Up`/`Down`), `db.go:10-16` embed, `migrate.go:13-65`, `admin/main.go:55-69`. The old `userdb/schema.sql` from `5b827f6` is deleted; exactly one `.sql` remains in the tree |
| Curl integration tests in `zarf/integration`, wired as `test-curl` | **Unmet** | `makefile:test-curl`, `api_test.sh`, `lib.sh` | Tests exist and are wired (`makefile:37-38`, `api_test.sh:19-30`), and assertion failures do exit 1 — but a preflight/tooling abort exits **0**. `lib.sh:154,188,192-200`. A target that cannot fail is not a gate |
| Print each curl request before running it | **Met** (two gaps) | `lib.sh:request` and every other curl call | `lib.sh:91-92` prints before the curl at `:98`. Unprinted: the preflight probe `lib.sh:177`, the cleanup DELETEs `userapp_test.sh:77-79` |
| Full CRUD for Users per `layered-architecture-types` | **Partial** | `userapp/{route,userapp,model}.go`, `userbus/{userbus,model}.go`, `userdb/{userdb,model}.go` | All five ops and all four converters present and correctly directioned (`userapp/model.go:30,63,110`; `userdb/model.go:21,34`). Gaps: primitive `int64` ID in Business (F5), uncofirmed pointer fields (F8), path parsing outside a converter (F9) |
| Integration tests split per domain | **Met** | `zarf/integration/*_test.sh` | `checkapp_test.sh`, `helloapp_test.sh`, `userapp_test.sh`, `mux_test.sh`, runner `api_test.sh:24-30`, shared `lib.sh` |

## Spec/AC weaknesses

- **S1 — "reply a 200 to any GET request" is ambiguous, and code and requirement disagree in writing.** Read literally this is a catch-all GET handler; the implementation registers only `GET /healthcheck` and `mux_test.sh:15-17` actively asserts `GET /does-not-exist` → 404. Neither document can settle a future dispute. Tighten to: *"`GET /healthcheck` returns 200 with `application/json` body `{"status":"OK"}`. Unregistered paths return 404; a wrong method on a registered path returns 405."*
- **S2 — no exit-code contract for `test-curl`.** "Link them in the make file" says nothing about what the target returns when the service is down, a tool is missing, or zero tests run — which is exactly why F1 shipped unnoticed. Tighten to: *"`make test-curl` exits 0 only if at least one assertion ran and all passed. It exits non-zero on any assertion failure, an unreachable service, a missing prerequisite (`curl`, `jq`), or zero assertions run. Skipped tests do not count as passes."* This single criterion would have caught both F1 and F6.
- **S3 — no empty/edge criteria for List Users.** "Render them as a big JSON for now" states nothing for zero users (`[]` vs `null` vs 404), ordering, or field/date formats. That absence is why the empty-DB path became a `skip` instead of an assertion. The code's choices are sound but unrequired: `[]` for empty, `ORDER BY user_id` (`userdb.go:108-109`), RFC3339 dates (`userapp/model.go:116-117`). Tighten to name all four.
- **S4 — "the CRUD for Users" leaves the entire failure surface unspecified.** No status codes, validation rules, duplicate-email behavior, partial-update semantics, or unknown-id behavior. The implementation picked sensible answers (201/200/204/400/404/409) and `userapp_test.sh` documents them retroactively — the tests are currently the only spec, so a refactor could change 409→500 and violate no criterion. Tighten to state each code, plus *"invalid `name`/`email` → 400 with a JSON array of `{field, error}` reporting all invalid fields at once"* and *"PUT is a partial update; omitted fields are left untouched."*

## Fix queue

### G4 Stop

**F1 — [Error Tripwire, Spec Cartographer] The integration harness reports success after a hard failure — `zarf/integration/lib.sh:145-154, 191-200`**

- Proof: `trap _exit_handler EXIT` ends in `summary`, whose final statements are `if [[ ${failed} -gt 0 ]]; then exit 1; fi` / `exit 0`. An explicit `exit` inside an EXIT trap **replaces** the script's exit status, so `preflight`'s `exit 1` at `:170` (missing `curl`/`jq`) and `:188` (service unreachable) are both overwritten. Reproduced directly:
  ```
  $ WAIT_SECONDS=1 ./zarf/integration/api_test.sh http://127.0.0.1:59999
  error: service not reachable at http://127.0.0.1:59999 after 1s
  0 passed, 0 failed
  $ echo $?
  0
  ```
  `lib.sh:18` sets `-u -o pipefail` but not `-e`, so a mid-script abort (unbound variable, a suite that dies while sourced) is masked the same way: the trap always re-derives the status from `failed`, and zero tests executed counts as a pass.
- Why it matters: `make test-curl` is green against a service that was never running. "Harness crashed" and "all tests passed" are indistinguishable. Any CI wiring this in gets a permanently passing gate. The assertion-failure path is fine — a real failed assertion does exit 1 — which is what makes this narrow defect so easy to miss.
- Patch direction: capture the incoming status first — `_exit_handler() { local rc=$?; …; summary "$rc"; }` — and have `summary` exit non-zero when `rc != 0` **or** `failed > 0`. Additionally treat `passed == 0 && failed == 0` as a failure so an empty run cannot pass. Have `preflight` set a flag rather than calling `exit` itself.
- Verify: `WAIT_SECONDS=1 ./zarf/integration/api_test.sh http://127.0.0.1:59999; echo $?` must be non-zero; a genuine all-pass run against a live service must still be 0; `mux_test.sh` against a mock returning 200 for `/does-not-exist` must stay at exit 1.

**F2 — [Security Sentinel] Every user CRUD route is unauthenticated and unauthorized — `app/domain/userapp/route.go:21-25`, `app/sdk/mux/mux.go:21-28`**

- Proof: `Routes` binds all five handlers with `api.Wrap(cfg.Log, …)` and nothing else; `api.Wrap` (`app/sdk/api/api.go:48-75`) calls the handler directly with no identity extraction, and `mux.WebAPI` wraps the `ServeMux` in no middleware. A repo-wide grep for `authoriz|authentic|bearer|jwt` across all `.go` files returns **0 matches** — verified independently. `userapp.user()` (`userapp.go:106-118`) resolves the target purely from the path integer, so there is no owner check either. The server binds `0.0.0.0:3000` by default (`api/services/api/main.go:44`).
- Exploit: `curl http://host:3000/users` dumps every user's name and email; `curl -X DELETE http://host:3000/users/1` deletes a user; `curl -X PUT -d '{"email":"attacker@x.com"}' http://host:3000/users/1` takes over a record. No credential or precondition needed.
- Assumption to confirm: that these endpoints are not deliberately public for this exercise step. Note the schema has no password/hash/role/tenant column and `userbus` has no password handling at all — a `users` domain with no identity concept anywhere. Nothing in the code contract defers auth to a later layer, so the gap is stated rather than assumed away.
- Patch direction: an authentication middleware in `app/sdk/mux` populating an identity in the context, plus an authorization check in each `userapp` handler — admin-only for `queryAll`/`create`/`delete`, owner-or-admin for `queryByID`/`update`, compared against `usr.ID` **before** any side effect. Requires identity/role columns in the schema.
- Verify: add no-credential cases to `zarf/integration/userapp_test.sh` asserting 401 on all five routes, and a wrong-owner case asserting 403 on `GET`/`PUT`/`DELETE /users/{other_id}`.

**F3 — [Harness Map] `make test` passes with zero Go tests; nothing gates this branch — `makefile:29-31`**

- Proof: `find . -name '*_test.go'` returns **0** results (verified independently); `go test ./...` prints `[no test files]` for all 16 packages and exits 0. The only real assertions live in `zarf/integration/*_test.sh`, which no `test` target invokes — `test-curl` is separate and needs a live service plus a migrated database, and per F1 cannot fail on its own preflight.
- Why it matters: `make test` returns a pass signal carrying no information. Every regression below ships silently. Compounding it, the branch's one honest gate (`test-curl`) is manual, requires a live stack, and is itself broken.
- Patch direction: add Go unit tests for the cheap pure logic — `business/types/{name,mail}` boundaries, the four converters, `errs`, and `api.Wrap` via `httptest`. Rename `test` → `test-unit` and add a `test-all` that also runs `test-curl`, or make `test` fail loudly while no tests exist.
- Verify: `make test` exercises named subtests instead of printing 16 `[no test files]` lines.

**F4 — [Harness Map] The only validation in the system is untested, and its real boundary is not what it looks like — `business/types/name/name.go:10`, `business/types/mail/mail.go:41`**

- Proof: `^[a-zA-Z][a-zA-Z0-9_ -]{2,20}$` — one leading character plus 2–20 more — accepts length **3 through 21**, not 3–20 as the shape suggests. `mail.Parse` delegates to `net/mail.ParseAddress`, which accepts the display-name form: `"Dave <dave@example.com>"` parses, and `Address.String()` (`mail.go:14-21`) returns only `dave@example.com`, silently discarding `Dave`. It also accepts `a@b` with no TLD. The shell suite exercises exactly one invalid name (`"a"`, `userapp_test.sh:121`) and one invalid email (`not-an-email`, `:133`) — no boundary case, no accept-side case.
- Why it matters: this validation guards a persisted unique column and is the only validation in the system. A one-character regex edit changes what the API accepts and nothing fails — including dropping the space from the character class, which would reject the very `"Dave Tester"` the suite uses everywhere else. The display-name case is worse: `POST /users` with `"email":"Dave <d@e.com>"` returns **201 with an email different from what the client sent**, and no test observes the rewrite.
- Patch direction: table-driven `Parse` tests in both packages — min/max/over-max length, leading digit, leading space or hyphen, empty; for mail, plain address, display-name form (assert the intended behavior explicitly, whichever it is), missing `@`, missing TLD, multiple addresses. Add the `MustParse` panic cases and the `Address{}.String() == ""` zero-value guard. Decide whether the regex bound should read `{2,19}`.
- Verify: `go test ./business/types/...`

### G3 Repair

**F5 — [Boundary Keeper] `userbus.User.ID` is a primitive `int64`; no strong ID type exists — `business/domain/userbus/model.go:12`**

- Proof: `SKILL.md:32` is MUST-level: Business models use "strong types for IDs, enums, classifications". `User.ID` is a bare `int64`, and the same bare `int64` is the `Storer`/`Business` parameter type (`userbus.go:23`, `userapp.go:107`, `userapp/model.go:102`). No `business/types/userid` package exists anywhere in the diff.
- Why it matters: any `int64` in scope — a row count, an index, another domain's id — is assignable to `User.ID` and passable to `QueryByID` with no compile error. This is the one field in the domain that crosses all three layers, and it is exactly the primitive-leakage case the rule exists to prevent.
- Stated assumption: `user_id` is `BIGSERIAL`, so the value is DB-generated and a `Parse` constructor is awkward on the create path. `SKILL.md:20-27` makes this your call — but it requires an explicit decision, not a default.
- Patch direction: add `business/types/userid` with `type ID struct{ value int64 }`, `Parse(string)`, `ParseInt(int64)`, `MustParse`, `String()`, `Int64()`. Then `User.ID userid.ID`, `Storer.QueryByID(ctx, userid.ID)`; `toDBUser` flattens with `.Int64()`, storage `toBusUser` parses, `fromBusUserResponse` emits `.Int64()`.
- Verify: `grep -rn 'int64' business/domain/userbus app/domain/userapp` returns only DB-row/response fields and the `userid` internals.

**F6 — [Doc Drift Check] `jqv` documents "prints nothing" but prints `null`, making a test's failure branch unreachable — `zarf/integration/lib.sh:121-125`**

- Proof: the comment claims it prints nothing "when the body is not valid json or the path is absent". Only stderr is discarded (`2>/dev/null`); with `--exit-status`, jq writes `null` to **stdout** and exits 1 for an absent path. Verified directly: `echo '{"a":1}' | jq --exit-status --raw-output '.foo' 2>/dev/null` prints `null`, exit 1. The invalid-JSON half of the claim does hold (exit 5, no stdout).
- Why it matters: two call sites gate on emptiness and are silently broken. `userapp_test.sh:104-110` does `USER_ID="$(jqv '.id')"; if [[ -n "${USER_ID}" ]]` — a 201 response missing `id` yields `USER_ID=null`, so `fail 'response carries an id'` is unreachable, the suite reports a false pass, and `cleanup_users` (`:74-80`) issues `DELETE /users/null`. The same pattern at `:21-24` would make `[[ null -gt 0 ]]` a bash arithmetic error.
- Patch direction: fix the code to match the comment — suppress stdout on non-zero jq exit — since two call sites already depend on the documented behavior. Then extend the comment to cover null/false values.
- Verify: `source lib.sh; RESP_BODY='{"a":1}'; v="$(jqv '.id')"; [[ -z "$v" ]] && echo ok`

**F7 — [Security Sentinel, Service Steward, Error Tripwire] Internal error text, including database connection details, is returned to clients — `app/sdk/api/api.go:101-109`, `app/domain/userapp/userapp.go:89,138-139`**

- Proof: `respondError` starts with a scrubbed body (`errs.Error{Message: http.StatusText(500)}`) but replaces it with `data = err` as soon as `errors.As(err, &st)` succeeds. `*errs.Error` implements `HTTPStatus()` (`errs.go:39-41`), and `busError`'s `default:` branch builds exactly such an error via `errs.Newf(500, "%s: %s", op, err)` — so the scrubbing branch is **dead code** for every error these handlers produce, and the fully wrapped chain `queryall: select: <pgx error>` is marshalled verbatim through the `json:"message"` tag.
- Exploit: request `GET /users` while Postgres is unreachable — `sqldb.Open` is lazy (`sqldb.go:45`), so connect failures surface at query time, and the pgx error text embeds the DB user, database name and host:port: `failed to connect to \`user=postgres database=postgres\` (localhost:5432)`. Other pg errors return SQLSTATE plus table and column names — a free schema map. `decode` (`userapp.go:123`) similarly returns `encoding/json` parser internals.
- Why it matters: reconnaissance handed to an unauthenticated caller (compounded by F2), and the scrubbing code reads as though 500s are sanitized when they never are.
- Patch direction: keep the detailed error in the log only (`api.go:54` already logs it). In `respondError`, retain the generic body whenever `statusCode >= 500`; or have `busError`'s default branch return `errs.Newf(500, "internal error")` and reduce `decode` to `"invalid json"`. Keep 4xx messages, which are client-actionable.
- Verify: point the service at a dead DB port and `curl -i http://localhost:3000/users` — the body must contain no `user=`, `database=`, host, or SQLSTATE, while the log line still carries the full chain.

### G2 Tighten

**F8 — [Boundary Keeper, Service Steward, Spec Cartographer] Pointer fields across three layers without the skill's required type confirmation — `app/domain/userapp/model.go:17,22-26`, `business/domain/userbus/model.go:29-33`**

- Proof: `SKILL.md:47` — "Prefer non-pointer types at every layer… reach for a pointer only when nothing else expresses the requirement, and say why"; `:56` requires surfacing pointers in the type confirmation; `:53` is directly on point for create requests: "if field absence is the only reason for a pointer, make the field required or default it to zero". The only justification recorded is a comment restating the mechanism ("A nil field means the existing value is left untouched").
- Assessment: the `Update*` pointers do carry real information (nil = untouched, correctly consumed at `userbus.go:58-68`) and partial update is the canonical justification — defensible, but it must be *recorded*. `NewUserRequest.Enabled *bool` is the weak case: absence is its only purpose, and `enabled BOOLEAN NOT NULL DEFAULT TRUE` (`00001_create_users_table.sql:6`) already encodes that default, so two mechanisms now express one rule.
- Patch direction: record the boundary type decisions for sign-off. For `NewUserRequest.Enabled`, either make it a plain `bool` or omit `enabled` from the INSERT so the column default applies.
- Verify: the confirmation exists in the review record; `POST /users` without `enabled` still yields the agreed value.

**F9 — [Boundary Keeper] Path-parameter parsing happens in a handler, not a named converter — `app/domain/userapp/userapp.go:106-110`**

- Proof: `strconv.ParseInt(r.PathValue("user_id"), 10, 64)` is inlined in `(*app).user` and fails with a bare `errs.Newf(400, …)`. `SKILL.md:69` is MUST-level: "All parsing/validation happens in App-layer `toBus<Type>`… on failure `fieldErrors.Add(field, err)`".
- Why it matters: `user_id` is the one App→Business crossing with no named converter, so it is the crossing that will drift. Its error shape also differs from the other converters — clients get two different failure formats for the same class of bad input.
- Patch direction: extract `toBusUserID(s string)` in `model.go` returning `errs.FieldErrors` keyed `"user_id"` (folds into F5).
- Verify: no `strconv.` call remains in `userapp.go`; `GET /users/abc` returns the same `{field, error}` shape as an invalid-email create.

**F10 — [Service Steward, Straight-Line Pass, Harness Map] `make tidy` runs `go mod vendor` in a repo with no `vendor/` — `makefile:10-12`**

- Proof: no `vendor/` directory exists, and `.gitignore:23` leaves `vendor/` commented out, so it is neither committed nor ignored. `go mod vendor` creates it, after which Go defaults to `-mod=vendor` for every subsequent build.
- Why it matters: a "tidy" target silently changes the module's build mode for every later command, resolving from an untracked directory no other developer or CI job has, and floods `git status` with thousands of untracked files. Any dependency change not re-vendored then breaks `go build ./...`.
- Patch direction: drop `go mod vendor` from `tidy`, or commit `vendor/` deliberately and add a CI sync check. Today's makefile straddles both.
- Verify: `rm -rf vendor && make tidy && git status --short` stays clean.

**F11 — [Spec Cartographer, Harness Map] The List-Users payload assertions silently skip on a fresh database — `zarf/integration/userapp_test.sh:26,46-49`**

- Proof: `GET /users` runs at `:16`, but every field/type/RFC3339 assertion is gated behind `[[ "${count:-0}" -gt 0 ]]` and otherwise falls through to `skip`, which by design (`lib.sh:61-66`) increments neither counter. The suite's own user is not created until `:86` — after the list assertions.
- Why it matters: on a clean database — the primary demo path after `make migrate` — the payload shape of "render them as a big JSON", the whole point of the original requirement, is unverified and the run still reports green. `skip` is honest on screen but invisible to the exit code.
- Patch direction: move the lifecycle `POST` block above the `GET /users` block so the list is guaranteed non-empty, and assert the created id appears in the collection. Reserve `skip` for genuinely unreachable cases.
- Verify: `make migrate-down && make migrate && make test-curl` — the list-shape assertions appear as passes, not skips.

**F12 — [Service Steward] `GET /users` is an unbounded full-table read — `business/domain/userbus/stores/userdb/userdb.go:100-113`**

- Proof: `SELECT … FROM users ORDER BY user_id` with no `LIMIT`/`OFFSET`; `Storer.QueryAll(ctx)` takes no paging arguments, `queryAll` accepts no query string, and `fromBusUsersResponse` materializes the whole set before marshalling.
- Why it matters: one request can pull the entire user table into memory and into a single JSON body. Retrofitting paging later changes the `Storer` interface, the bus API, and the response shape — breaking clients.
- Patch direction: decide the contract now — either add `page`/`rows` (or keyset-on-`user_id`) through `Storer.Query`, or state the bound with an explicit `LIMIT`.
- Verify: seed ~100k rows and time `curl -s localhost:3000/users | wc -c`.

**F13 — [Security Sentinel] Insecure database defaults: built-in credentials and TLS off — `app/sdk/env/db.go:9-15`**

- Proof: `DBConfig()` defaults `DB_USER`/`DB_PASSWORD` to `postgres`/`postgres` and `DB_DISABLE_TLS` to `true`; `env.String` silently falls back whenever the variable is unset, and `sqldb.Open` then sets `sslmode=disable` (`sqldb.go:28-34`). Both binaries use it.
- Why it matters: a deploy missing `DB_DISABLE_TLS=false` sends credentials in cleartext on every connection; one missing `DB_PASSWORD` tries the default `postgres` superuser password. These are developer conveniences — which is exactly why fail-open matters, since the same code path ships.
- Patch direction: make credentials and host mandatory with no default (a `MustString` that errors when unset), and invert the flag to `DB_ENABLE_TLS` defaulting secure (`DisableTLS: !Bool("DB_ENABLE_TLS", true)`) so opting out is explicit. Keep dev values in a local `.env` or makefile variable, not in Go.
- Verify: `env -u DB_PASSWORD -u DB_DISABLE_TLS go run api/tooling/admin/main.go migrate-status` must fail with a configuration error rather than connect.

**F14 — [Security Sentinel] Unbounded request body on POST and PUT — `app/domain/userapp/userapp.go:121-127`**

- Proof: `json.NewDecoder(r.Body).Decode(v)` with no `http.MaxBytesReader` and no size cap; `main.go:78-85` sets read/write/idle timeouts but no body limit, and neither `api.Wrap` nor `mux.WebAPI` adds one.
- Why it matters: a multi-gigabyte JSON string in `"name"` is buffered as the decoder builds the request struct. The 5s `ReadTimeout` bounds a slow trickle but not a fast large body on loopback or LAN; a few concurrent requests exhaust process memory. Unauthenticated, per F2.
- Patch direction: in `decode`, `r.Body = http.MaxBytesReader(w, r.Body, 1<<20)` — or cap it in mux-level middleware to cover future handlers — and enable `dec.DisallowUnknownFields()` while there.
- Verify: a 50 MB `name` value must return 413 with flat process RSS.

**F15 — [Error Tripwire, Boundary Keeper, Doc Drift, Clone Hunter] `isUniqueEmailViolation` matches any unique violation — `business/domain/userbus/stores/userdb/userdb.go:166-172`**

- Proof: the doc comment says "on the email column", but the body tests only `pgErr.Code == uniqueViolation`; `ConstraintName` is never inspected. (`23505` is confirmed the correct `unique_violation` SQLSTATE.) Correct today only because `email` is the sole non-PK unique constraint (`00001_create_users_table.sql:5`) and `user_id` is never supplied by the INSERT or UPDATE.
- Why it matters: the result maps to `ErrUniqueEmail` → 409 "email is not unique" (`userapp.go:135-136`). The first added unique index reports the wrong field to clients, and the comment is what will tell the next maintainer this is already handled.
- Patch direction: match `pgErr.ConstraintName` against a named email constraint (name it in the migration), letting other `23505` codes fall through; or reword the comment to make the assumption visible.
- Verify: unit-test with `&pgconn.PgError{Code:"23505", ConstraintName:"users_pkey"}` and assert false.

**F16 — [Clone Hunter] The six-column list is written out four times — `business/domain/userbus/stores/userdb/userdb.go:41,68,105,123`**

- Proof: the literal `user_id, name, email, enabled, date_created, date_updated` appears four times — twice as `RETURNING`, twice as `SELECT`. The `dbUser` struct (`userdb/model.go:12-19`) is the single declaration of the shape, and none of the four queries is tied to it.
- Why it matters: the classic add-a-column bug. Adding a field means editing four string literals; miss one and `Create` silently returns a zero-valued field while `QueryByID` returns the real one — no compile error, because `StructScan` tolerates a narrower result set.
- Patch direction: one `const userColumns = "…"` interpolated into all four. Deliberately not struct-tag reflection — explicit SQL is the house style; just write the list once.
- Verify: `rg -c 'date_updated' business/domain/userbus/stores/userdb/userdb.go` drops from 5 to 2 — one `const` plus the `date_updated = NOW()` in the UPDATE `SET` clause, which is a genuinely separate use and must stay. The field-shape assertions must still pass.

**F17 — [Straight-Line Pass] Four exported functions with zero callers — `app/sdk/errs/errs.go:18,53`, `business/types/mail/mail.go:47`, `business/types/name/name.go:47`**

- Proof: `grep -rn --include='*.go' -E 'errs\.New\(|errs\.NewFieldErrors|name\.MustParse|mail\.MustParse' .` returns nothing, and there are no test files that could be the intended consumers.
- Why it matters: unexercised exported surface drifts silently. `errs.New` also returns `*Error` while `respondError` builds a value `errs.Error{}` — two shapes for one concept, and only the unused one is a pointer.
- Patch direction: delete `errs.New` and `errs.NewFieldErrors`; keep the two `MustParse` functions only if the tests that use them land with F4, otherwise re-add them with their first caller.
- Verify: `go build ./... && go vet ./...` after deletion.

**F18 — [Service Steward, Boundary Keeper, Clone Hunter] `errs.FieldErrors` has no `Add`; the accumulate idiom is hand-rolled four times — `app/domain/userapp/model.go:35,40,70,80`**

- Proof: each site writes `append(fieldErrors, errs.FieldError{Field: "name", Err: err.Error()})`. `SKILL.md:69,96-98` specifies `fieldErrors.Add(field, err)`; `errs` exposes only `NewFieldErrors`, which is itself unused (F17). The field-name literals `"name"` and `"email"` each appear twice in separate functions, and `userapp_test.sh:140` and `:207` assert on those keys independently — so a half-rename passes one test and fails the other.
- Patch direction: add `func (fe *FieldErrors) Add(field string, err error)` and convert the four sites.
- Verify: `grep -rn "append(fieldErrors" --include='*.go' .` returns nothing; the "reports every invalid field at once" test (`userapp_test.sh:143-152`) still returns both `name` and `email` — that test guards the accumulate-don't-early-return behavior any refactor here could break.

**F19 — [Boundary Keeper] `Storer.Create` takes a full `User`, so the write port cannot express "not yet persisted" — `business/domain/userbus/userbus.go:19,41-53`**

- Proof: `Create(ctx, usr User) (User, error)`; `Business.Create` leaves `ID`, `DateCreated`, `DateUpdated` zero, `toDBUser` flattens those zeros, and the INSERT (`userdb.go:35-41`) silently ignores them, binding only `:name, :email, :enabled`.
- Why it matters: the store's input type claims three fields it does not read. A future author adding `user_id`/`date_created` to the INSERT column list writes a zero id and zero timestamp with no compile error. `NewUser` already exists as the right shape.
- Patch direction: `Storer.Create(ctx, nu NewUser) (User, error)` with a `toDBNewUser` carrying only `name/email/enabled` — matching `SKILL.md:72`.
- Verify: `toDBUser` no longer appears on the create path; `POST /users` still returns a server-assigned id and timestamps.

**F20 — [Harness Map] Untested CRUD behaviors that a plausible edit would break silently**

- `userbus.Update` partial merge (`userbus.go:57-76`): the only successful PUT sends `{"name":…}` (`userapp_test.sh:174-187`). No email-only, `enabled`-only, or multi-field update exists, and `Update`'s own 409-on-duplicate path (`userdb.go:72-74`) is never hit. Swapping `usr.Email = *uu.Email` for `usr.Name`, or dropping the `Enabled` block entirely, passes every existing test — `enabled` is asserted only as "preserved" (`:186`), which is indistinguishable from a merge that ignores it.
- Explicit `"enabled": false` on create (`userapp/model.go:47-50`): never sent by any test; only the omitted-field default is asserted (`:94`). Collapsing the pointer to `bool` — the exact simplification F8 contemplates — would make `false` silently become `true` with every test still green.
- The error envelope: every non-field-error assertion checks only the status code (404 at `:237`, 409 at `:159`, 400 at `:255`). The `{"message": …}` shape is entirely unasserted, as is `respondError`'s non-`Statuser` fallback. Two different error shapes ship on the same 400 status — a bare JSON array for validation, an object for everything else — and nothing pins either.
- Converter error paths (`userdb/model.go:36-43`): `toBusUser`'s two failure branches (a stored row whose name or email no longer parses) are unreachable through the API and would surface as an opaque 500 on `GET /users`. The `.UTC()`/`.In(time.Local)` round-trip is unasserted.
- Patch direction: a `userbus.Update` table test against a fake `Storer` (each field alone, all three, all-nil); `toBusNewUser` tests for absent/false/true; `httptest` tests of `api.Wrap` for `*errs.Error`, `errs.FieldErrors`, and a plain `errors.New` (asserting the internal message does **not** appear); a `toDBUser`→`toBusUser` round-trip using the existing `Equal` methods.
- Verify: `go test ./app/... ./business/...`

**F21 — [Clone Hunter] Duplicated idioms that will drift**

- `env.Int`/`env.Bool`/`env.Duration` (`app/sdk/env/env.go:21-52`) are one function written three times, byte-identical apart from the parse call. The shared policy is "silently swallow a malformed value and use the default" — a policy worth changing, and today that is three edits with three chances to miss one. Extract `func parse[T any](key string, def T, fn func(string) (T, error)) T`.
- The parse-and-accumulate idiom in `userapp/model.go:33-41` and `:66-84` (value and pointer flavours). Not yet drifted — the asymmetry is required by partial update — but the two functions must agree on the client-visible field names with nothing enforcing it. Extract `parseField`/`parseOptional` generic helpers.
- The response-shape assertion block in `userapp_test.sh:29-37` and `:96-102` is the executable spec for `UserResponse`, duplicated. Add a field and only one copy gets it, leaving the collection and create endpoints held to different contracts. Extract `assert_user_shape` into `userapp_test.sh` (not `lib.sh`, which should stay domain-agnostic).

### G1 Polish

- **Dead empty-user guard** — `app/domain/helloapp/helloapp.go:20-23`. Verified against `net/http` on go1.26.4: `GET /hello/{user}` gives `/hello/dave` → 200, `/hello/` → 404, `/hello` → 404. A wildcard never binds an empty segment, so `user == ""` is unreachable and the `errs` import exists only to serve it. `helloapp_test.sh:28-31` asserts the surrounding 404, which makes the guard *look* covered. Delete it and the import.
- **`StatusData.HTTPStatus` is unreachable** — `app/sdk/api/api.go:42`. `Wrap`'s type switch lists `case StatusData` before `case Statuser`, and Go takes the first match, so the method can never be called. Delete it.
- **`queryAll` bypasses `busError`** — `app/domain/userapp/userapp.go:89`. Every other handler routes through it. No behavioral difference today since `QueryAll` surfaces no sentinels, but it is the one handler that cannot map a future one. Also the source of the doubled `"queryall: queryall: …"` prefix.
- **`errors.As` where Go 1.26 `errors.AsType` reads better** — `userdb.go:168-172` becomes `pgErr, ok := errors.AsType[*pgconn.PgError](err)`. This does **not** apply to `api.go:106`: `AsType[E error]` constrains `E` to `error` and `Statuser` does not embed it.
- **`env.String` is `cmp.Or`** — `app/sdk/env/env.go:37-43`, modulo the set-but-empty case.
- **Redundant import alias** — `api/tooling/admin/main.go:11` aliases `db "…/zarf/schemas/db"` where the package is already `db`.
- **`toBusUsers` assigns through a pre-declared `err`** — `userdb/model.go:57-66`. `b, err := toBusUser(db)` then assign reads in one pass.
- **Store `Delete` reports success for a row that never existed** — `userdb.go:83-99` ignores `RowsAffected`. Safe over HTTP only because `userapp.delete` pre-loads via `a.user` (`userapp.go:73`) — an undocumented compensating control that is also a TOCTOU under concurrent deletes, and any other caller of the public `userbus.Delete` gets a silent no-op. A comment at minimum; `RowsAffected` → `ErrNotFound` properly.
- **`/healthcheck` never touches the database** — `checkapp/checkapp.go:16-18` always returns `{"status":"OK"}` while `sqldb.StatusCheck` (`sqldb.go:60`) already exists. Liveness-only is defensible, but `lib.sh:177` gates the whole suite on it, so the harness declares a DB-less service "ready" and the user suites then fail with 500s that look like product bugs.
- **Every handler error logs at `Error` level** — `app/sdk/api/api.go:105`. 400-class validation failures generate the same error-level noise as 500s.
- **`update` returns 404 before 400** — `userapp.go:53-61` loads the user before decoding, so a malformed body against an unknown id reports 404 rather than the 400 the body deserves.
- **`request`'s echoed curl line is not copy-pasteable** — `lib.sh:81-92` prints `"${cmd[*]}"`, which drops all quoting, so any body containing a space (every POST/PUT in the suite) prints a line that fails differently from the test. Either print with `printf '%q '` over `"${cmd[@]}"` or drop the "copy-pasted as-is" claim.
- **`make test-curl`'s comment names only `run`, not `migrate`** — `makefile:33-36`. `StatusCheck` only pings and runs `SELECT true`, so `make run` succeeds against a database with no `users` table; every user test then 500s while preflight reports "ready".
- **`StatusCheck`'s doc omits the retry loop and its 1s default deadline** — `sqldb.go:56-63`. Both callers pass a deadline-free context, so startup gives the database exactly 1s and `make run` against a still-starting Postgres reads as a config error.
- **`mail.Parse`'s doc hides display-name acceptance and normalization** — `mail.go:33-42`. "The rules for an email address" reads as addr-spec; contrast `name.Parse`, whose regex genuinely is strict. See F4.
- **Two unprinted curl calls** — `lib.sh:177` (preflight probe) and `userapp_test.sh:77-79` (cleanup DELETEs), against the "print the request before running it" requirement.
- **`checkapp`/`helloapp` boilerplate is byte-identical** — recorded only so it is not mistaken for an unreviewed clone. Do **not** unify: each `Config` is intentionally per-domain, as `userapp/route.go:12-15` proves by adding `UserBus`.

## Open checks

- **F2 needs your ruling.** Whether public unauthenticated user CRUD is acceptable for this exercise step is the one finding that turns on intent. Everything else in the report stands either way.
- **No SAST or secret scanner is available** — gosec, gitleaks, trufflehog, semgrep, codeql, osv-scanner and trivy are all absent. The secrets and injection conclusions rest on manual reading plus targeted greps. Git *history* was not swept for secrets; only diff content was.
- **No clone detector is available** — `dupl`, `jscpd` and PMD `cpd` are all absent, so F16/F21 are inspection-based. `dupl -threshold 40 ./...` in the `test` target would make that rung mechanical.
- **Nothing was confirmed dynamically against a live service or database.** F1, F6 and the exit-code behavior were reproduced directly; F7's leak text and F14's memory exhaustion are derived from the code path and pgx/`encoding/json` error formats, not from an observed response. Both are cheap to confirm with the Verify steps.
- **`name.Parse` / `mail.Parse` vs `Parse<Type>`** — `SKILL.md:74-80` makes `Parse<Type>`/`MustParse<Type>` a MUST. Read literally these should be `ParseName`/`ParseAddress`; the skill's example (`types.ParseOwnerID`) assumes a single shared `types` package where the suffix is the only disambiguator, whereas here each type has its own package and `name.Parse` reads identically at the call site. Surfacing for a ruling, not asserting a violation — either rename or add the carve-out to the skill.
- **The `ExtBusiness`/`Extension` seam is absent from `userbus`** — `NewBusiness` returns a concrete `*Business` (`userbus.go:33`) and both `mux.Config` and `userapp.Config` hold `*userbus.Business`. Not filed as a violation: `business-layer-extensions` triggers when a cross-cutting concern is added, and this diff adds none. Worth deciding now whether the seam lands with the scaffolding, since the first extension will otherwise require edits in App files — which is precisely what the seam exists to prevent.
- **Migrations have no automated coverage** — the `-- +goose Down` (`DROP TABLE users`) is never executed by anything, so `make migrate-down` is a destructive developer operation with no verification that it works or that a subsequent `migrate` restores the schema. Treating "migrations are dev-only tooling" as *not* a reason to skip this; say so explicitly if that is the deliberate position.
- **`toBusUser` re-parses stored values through `name.Parse`** (`userdb/model.go:34-43`) while the column is plain `TEXT` with no length constraint. Any row written outside this service that fails the regex makes `GET /users` return 500 for the whole collection (`toBusUsers` aborts on the first bad row). Correct fail-loud behavior, but it assumes this service is the only writer — confirm.
- **Timestamps are storage-owned** — `DEFAULT NOW()` on insert, `NOW()` in the UPDATE (`userdb.go:64`); Business never sets them. Deliberate, or should Business own the clock for testability?
- **`admin`'s migrations share a hard 1-minute context** (`api/tooling/admin/main.go:42`). A migration exceeding it is cancelled and rolled back and the command does exit non-zero, but the operator sees `context deadline exceeded` rather than a SQL error. Confirm 60s is the intended ceiling.
- **Concurrency is untested throughout.** `userapp_test.sh` is strictly sequential; the Delete TOCTOU is the concrete instance.
- **Refuted and dropped, so they are not re-raised:** `api/services/api/main.go` vs `api/tooling/admin/main.go` are *not* clones — the identical 118-line counts are coincidence, and the only shared sequence (8 lines of `sqldb.Open`/`StatusCheck`/`Close`) is already deduped behind `env.DBConfig`. `mail.go` vs `name.go` share only the value-type contract; every body differs, and the nil guard `mail` has and `name` lacks is *required* by the pointer field, not drift. The `zarf/integration/*.sh` suites reimplement no helper from `lib.sh`. `lib.sh`'s `local rc=$?` (`:103`) is correct — `$?` expands before `local` runs. The `\n#` sentinel parsing was verified against a live server for both 200 and 204 with no field shift. `23505` is the correct `unique_violation` code. Empty lists marshal as `[]`, never `null`, because both converters use `make`. The old `userdb/schema.sql` really is deleted. Goose really does discover the embedded migration (`v=1 00001_create_users_table.sql`). `userdb.Delete` mapping no postgres errors is correct, not drift — an INSERT cannot return `ErrNoRows`. `lib.sh` being mode 0644 while the suites are 0755 is correct, since it is only sourced.

## Suggested verification

```sh
go build ./... && go vet ./...                      # currently clean
go test ./... -count=1                              # currently 16× "no test files" — see F3
WAIT_SECONDS=1 ./zarf/integration/api_test.sh http://127.0.0.1:59999; echo $?   # must be non-zero after F1
make migrate-down && make migrate && make test-curl # list assertions must pass, not skip (F11)
go run golang.org/x/vuln/cmd/govulncheck@latest ./... # two go1.26.4 stdlib advisories; bump to 1.26.5+
```

Re-run Security Sentinel, Error Tripwire and Harness Map after the G4 fixes.
