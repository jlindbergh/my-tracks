# my-tracks
A tool for visualizing, filtering and viewing GPS located activities on a map

See [docs/tech-stack.md](docs/tech-stack.md) for the proposed tech stack and architecture.

## Running locally

Requires Docker with Compose. Nothing else needs to be installed on the host.

```bash
docker compose up --build
```

Then open http://localhost:5173 and import GPX files with the button, or drag and drop them onto it.

To batch-import a folder, put your files in [samples/](samples/) (subfolders are fine) and run:

```bash
docker compose run --rm backend import
```

Running it again is safe; see [Duplicate detection](#duplicate-detection).

Files without a GPS track (indoor rides, treadmill runs, gym sessions) aren't imported, since there is nothing to show on the map. The UI and the batch command report how many were skipped and their activity types. Points at 0,0 count as missing a GPS fix.

### Duplicate detection

Each import is checked against the user's own activities (never other users'):

- **Already imported:** a byte-identical file (same SHA-256) is always skipped.
- **Possible duplicate:** the file is *not* imported if an existing activity starts within **2 seconds** of it, and its distance and duration are both within **2%** of the larger value. Duration is skipped when either side has no timestamps.

GPS points aren't compared, because Garmin, Strava, Runkeeper etc. record and resample them differently. Files without timestamps can only be detected as already imported.

Possible duplicates can still be imported on purpose. The UI lists them side by side with the matching activity, explains the rules, and lets the user tick the ones to import anyway. The batch command takes `-allow-duplicates`.

The tolerances are the constants `DuplicateStartTolerance` and `DuplicateRelativeTolerance` in [backend/internal/importer/duplicates.go](backend/internal/importer/duplicates.go). The frontend reads them from `GET /api/import/rules`, so changing them there also updates what users are told.

### Services

| Service    | Module                     | Host port        | Notes |
|------------|----------------------------|------------------|-------|
| `frontend` | [frontend/](frontend/)     | `127.0.0.1:5173` | Vite dev server (React + MapLibre GL JS), proxies `/api` to the backend |
| `backend`  | [backend/](backend/)       | `127.0.0.1:8080` | Go API: auth, GPX import, tile proxy |
| `martin`   | [mapserver/](mapserver/)   | not published    | Martin vector tile server |
| `db`       | [postgis/](postgis/)       | `127.0.0.1:5433` | PostgreSQL 17 + PostGIS (h3-pg installed, not enabled) |

Connect with psql: `psql postgres://mytracks_owner:dev-owner-pw@localhost:5433/mytracks`.
Defaults can be overridden with environment variables or a `.env` file (see [compose.yaml](compose.yaml)).

### Resetting the database

The SQL in [postgis/init/](postgis/init/) only runs when the database volume is first created.
After changing it, recreate the volume (this deletes all imported data):

```bash
docker compose down -v
```

## How users are kept apart

A user must never see another user's tracks. The data path is built around that:

1. The browser only talks to the backend. Martin has no published port and no authentication of its own.
2. The backend resolves the user from the session and calls Martin with that user's id. Nothing from the client's query string is forwarded.
3. `activities_tiles()` in PostGIS returns no tile without a valid `user_id`, and only that user's activities with one.
4. Martin's database role can execute that function and nothing else. It cannot read the tables directly.

**Login does not exist yet.** With `AUTH_MODE=dev` every request is treated as `DEV_USER_EMAIL`, and the backend refuses to start in any other mode. OIDC (Zitadel/Strava) is the next step before this runs anywhere but localhost.

## Basemap

The default basemap is [OpenFreeMap](https://openfreemap.org/) ("Liberty" style). It's free with no API key or usage limits. The data is © OpenStreetMap contributors under the ODbL and the schema is OpenMapTiles, so attribution is required. The style shows it in the map's attribution control.

## Development

```bash
# Backend tests (no local Go needed). With the stack running, this includes
# the database tests; without TEST_DATABASE_URL they are skipped.
docker run --rm --network my-tracks_default \
  -e TEST_DATABASE_URL=postgres://mytracks_app:dev-app-pw@db:5432/mytracks \
  -v "$PWD/backend":/src -w /src golang:1.25-alpine go test ./...

# Frontend type check, with Node installed locally
cd frontend && npm install && npm run typecheck
```
