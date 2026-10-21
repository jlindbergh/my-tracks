# TODO

## Database migrations (soon)

The SQL in [postgis/init/](postgis/init/) only runs when the database volume is first created, so every schema or tile-function change needs `./dev.sh reset-db`, which deletes all data. That stops working once there is data worth keeping.

- Pick a migration tool (for example goose, golang-migrate or tern) and run migrations from the backend or a separate step in `dev.sh`.
- Turn the current init scripts into the first migrations. Roles and passwords (`010_roles.sh`) probably stay an init step.
- The simplification tolerances in the `geom_mid`/`geom_lo` generated columns are part of the schema, so tuning tiles will also need a migration.

## Generate TypeScript types from the Go API types

The API types in [frontend/src/api.ts](frontend/src/api.ts) are written by hand to mirror the Go structs (`activity.Facts`, `activity.Summary`, `activity.Comparison`, `importer.Result` and its statuses, the `/api/config` types), so they can drift apart without any error.

- Generate them from the Go structs, for example with [tygo](https://github.com/gzuidhof/tygo) added as a Go tool in `backend/go.mod`.
- Add `./dev.sh gen`, run in a Go container since nothing is installed on the host, and make `./dev.sh test` fail when the generated files are out of date.
- Check how `*time.Time`, pointers and embedded structs come out; struct tags may be needed.
