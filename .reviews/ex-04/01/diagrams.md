# Feature Diagrams — Users CRUD service (branch `ex-04`)

Drawn from the code as it stands on `main...HEAD`, not from the intended design.

## 1. Types across the three layers

Each layer owns its own shape; the arrows are the named converters.

```mermaid
flowchart LR
  subgraph APP["App — primitives only"]
    NUR["NewUserRequest<br/>Name string<br/>Email string<br/>Enabled *bool"]
    UUR["UpdateUserRequest<br/>Name *string<br/>Email *string<br/>Enabled *bool"]
    RES["UserResponse<br/>ID int64<br/>Name string<br/>Email string<br/>Enabled bool<br/>DateCreated string<br/>DateUpdated string"]
  end

  subgraph BUS["Business — strong types"]
    NU["NewUser<br/>Name name.Name<br/>Email mail.Address<br/>Enabled bool"]
    UU["UpdateUser<br/>Name *name.Name<br/>Email *mail.Address<br/>Enabled *bool"]
    U["User<br/>ID int64 ⚠<br/>Name name.Name<br/>Email mail.Address<br/>Enabled bool<br/>DateCreated time.Time<br/>DateUpdated time.Time"]
  end

  subgraph DB["Storage — db tags"]
    DBU["dbUser (unexported)<br/>user_id, name, email,<br/>enabled, date_created, date_updated"]
    TBL[("users table<br/>BIGSERIAL PK<br/>email UNIQUE<br/>all columns NOT NULL")]
  end

  NUR -->|toBusNewUser| NU
  UUR -->|toBusUpdateUser| UU
  U -->|fromBusUserResponse| RES
  NU --> U
  UU --> U
  U -->|toDBUser| DBU
  DBU -->|toBusUser| U
  DBU <--> TBL
```

`⚠ User.ID int64` is finding **F5**: the one field crossing all three layers is a bare primitive inside Business, where the layering rules require a strong type. Every other crossing is clean — no strong type reaches App JSON (both `name.Name` and `mail.Address` implement `MarshalText`, so a leak would have been silent, but `fromBusUserResponse` flattens each field explicitly), and `dbUser` is unexported so no row escapes the store.

## 2. Client contract

```mermaid
flowchart TD
  C(["client"]) --> MUX["http.ServeMux<br/>app/sdk/mux"]

  MUX --> H1["GET /healthcheck → 200 {status:OK}"]
  MUX --> H2["GET /hello/{user} → 200 {message}"]
  MUX --> H3["GET /users → 200 [ ]"]
  MUX --> H4["GET /users/{user_id} → 200 | 400 | 404"]
  MUX --> H5["POST /users → 201 | 400 | 409"]
  MUX --> H6["PUT /users/{user_id} → 200 | 400 | 404 | 409"]
  MUX --> H7["DELETE /users/{user_id} → 204 | 400 | 404"]

  H3 -.-> NOAUTH{{"no authentication<br/>or authorization<br/>on any route — F2"}}
  H5 -.-> NOAUTH
  H6 -.-> NOAUTH
  H7 -.-> NOAUTH

  style NOAUTH fill:#c0392b,color:#fff
```

Two distinct error envelopes ship on the same 400 status: validation failures return a bare JSON **array** of `{field, error}`, while everything else returns an **object** `{"message": …}`. Neither shape is asserted by any test outside the field-error case (F20).

## 3. Request flow — POST /users

```mermaid
sequenceDiagram
  participant C as client
  participant W as api.Wrap
  participant A as userapp.create
  participant B as userbus.Create
  participant S as userdb.Create
  participant P as postgres

  C->>W: POST /users {name, email}
  W->>A: handler(ctx, r)
  A->>A: decode(r) — no body size limit (F14)
  A->>A: toBusNewUser — name.Parse, mail.Parse<br/>accumulates all field errors
  alt invalid input
    A-->>W: errs.FieldErrors
    W-->>C: 400 [{field, error}, …]
  else valid
    A->>B: Create(ctx, NewUser)
    B->>B: build User, zero ID/timestamps
    B->>S: Create(ctx, User)
    S->>S: toDBUser — ID and timestamps<br/>flattened but never read (F19)
    S->>P: INSERT … RETURNING (6 columns, 1 of 4 copies — F16)
    alt SQLSTATE 23505
      P-->>S: unique violation
      S-->>B: ErrUniqueEmail
      B-->>A: wrapped
      A->>A: busError → 409
      A-->>C: 409 {"message":"email is not unique"}
    else ok
      P-->>S: row
      S->>S: toBusUser — re-parses stored values
      S-->>B: User
      B-->>A: User
      A->>A: fromBusUserResponse
      A-->>W: UserResponse (201 via StatusData)
      W-->>C: 201 {id, name, email, …}
    end
  end
```

## 4. Error translation, and where it leaks

```mermaid
flowchart TD
  E["store / bus error<br/>(wrapped: 'queryall: select: pgx …')"] --> BE["userapp.busError"]
  BE -->|ErrNotFound| C404["errs.Newf(404)"]
  BE -->|ErrUniqueEmail| C409["errs.Newf(409)"]
  BE -->|default| C500["errs.Newf(500, '%s: %s', op, err)<br/>⚠ embeds the full wrapped chain"]

  C404 --> RE["api.respondError"]
  C409 --> RE
  C500 --> RE

  RE --> SAFE["data = errs.Error{Message: StatusText(500)}"]
  SAFE --> CHK{"errors.As(err, &Statuser)"}
  CHK -->|"always true for *errs.Error"| OVER["data = err → body carries<br/>DB user, database name, host:port"]
  CHK -->|"never reached in practice"| DEAD["generic body<br/>(dead code)"]

  OVER --> CL(["client"])
  DEAD -.-> CL

  style OVER fill:#c0392b,color:#fff
  style DEAD stroke-dasharray: 5 5
```

This is finding **F7**. The scrubbing branch reads as though 500 responses are sanitized; because every error these handlers produce satisfies `Statuser`, it is unreachable and the full chain is marshalled to the client.

## 5. Integration harness control flow — why it always exits 0

```mermaid
flowchart TD
  START(["api_test.sh"]) --> SRC["source lib.sh<br/>set -u -o pipefail (no -e)"]
  SRC --> TRAP["trap _exit_handler EXIT"]
  TRAP --> PRE["preflight"]

  PRE -->|"curl/jq missing"| EX1["exit 1"]
  PRE -->|"service unreachable"| EX1
  PRE -->|ok| SUITES["run 4 domain suites"]

  SUITES --> ASSERT["assertions increment<br/>passed / failed / skipped"]
  ASSERT --> NAT(["script ends"])

  EX1 --> TRAPRUN["EXIT trap fires"]
  NAT --> TRAPRUN
  TRAPRUN --> SUM["summary"]
  SUM --> Q{"failed > 0 ?"}
  Q -->|yes| R1["exit 1 ✓"]
  Q -->|"no — including<br/>0 tests run"| R0["exit 0 ⚠<br/>overrides preflight's exit 1"]

  style R0 fill:#c0392b,color:#fff
```

Finding **F1**, reproduced: an `exit` inside an EXIT trap replaces the script's status, so a preflight abort with zero tests executed reports success. Separately, `skip` counts as neither pass nor fail (F11), so the List-Users payload assertions vanish from the exit code on an empty database.

## 6. Migration and startup wiring

```mermaid
flowchart LR
  SQL["zarf/schemas/db/<br/>00001_create_users_table.sql<br/>+goose Up / +goose Down"] -->|go:embed| DBPKG["zarf/schemas/db"]
  DBPKG --> MIG["business/sdk/migrate<br/>Up / Down / Status"]
  MIG --> ADMIN["api/tooling/admin<br/>migrate | migrate-down | migrate-status"]
  ADMIN --> MK1["makefile: migrate targets"]

  ENVV["app/sdk/env<br/>DB_* defaults ⚠ postgres/postgres, TLS off"] --> SQLDB["business/sdk/sqldb.Open<br/>sqlx + pgx"]
  SQLDB --> ADMIN
  SQLDB --> API["api/services/api<br/>StatusCheck → mux → ListenAndServe :3000"]
  API --> MK2["makefile: run"]

  style ENVV fill:#e67e22,color:#fff
```

The schema move is complete — exactly one `.sql` file exists in the tree and goose discovers it as `v=1`. The orange node is finding **F13**: unset `DB_*` variables fail open to well-known credentials with TLS disabled rather than failing loudly.
