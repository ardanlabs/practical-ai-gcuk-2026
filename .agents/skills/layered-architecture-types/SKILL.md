---
name: layered-architecture-types
description: <skill description>
---

# Layered Architecture Types

Data crosses three layers: App, Business, Storage.

Strong types from `business/types` live only in Business; the edges use primitives.
Every crossing goes via a converter - never assign to a boundary directly.

## Layers

| Layer    | Package                | Types                       | Struct                             |
| -------- | ---------------------- | --------------------------- | ---------------------------------- |
| API      | `app/domain/*app`      | primitives                  | `<Type>Request` / `<Type>Response` |
| Business | `business/domain/*bus` | strong (`business/types/*`) | model type                         |
| Storage  | `business/domain/*bus/stores/*db`       | native DB types             | `db<Type>` row                     |

Non-domain support code lives in `app/sdk/*` (e.g. `app/sdk/errs`, `app/sdk/api`).

## Crossings

| Path  | From     | Converter                | To       | Defined in    | Can fail?                |
| ----- | -------- | ------------------------ | -------- | ------------- | ------------------------ |
| write | API      | `toBus<Type>`            | Business | App layer     | yes — `errs.FieldErrors` |
| write | Business | `toDB<Type>`             | Storage  | Storage layer | no                       |
| read  | Storage  | `toBus<Type>`            | Business | Storage layer | yes — `error`            |
| read  | Business | `fromBus<Type>Response`  | API      | App layer     | no                       |

Crossing *into* strong types can fail; crossing *out* cannot - the value is already valid.

The two `toBus<Type>` rows are different functions in different packages: the App one takes a
`<Type>Request`, the Storage one a `db<Type>`.