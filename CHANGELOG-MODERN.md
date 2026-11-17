# Hammond Modern changes

## UI
- Responsive persistent sidebar with mobile drawer, route highlighting, theme toggle and admin navigation.
- Dashboard heading, vehicle summary, responsive metric cards, and German UI labels.
- German locale display formatting for fuel quantity and odometer in dashboard vehicle cards.
- Fixed vehicle store watcher argument ordering; corrected stats fallback currency and added error handling.
- Existing forms, vehicle details, and API logic remain intact, with global light/dark styling.

## Build / deployment
- New multi-stage Dockerfile using Go 1.25, Node 22 and Alpine 3.22 base images (not yet validated by a full build).
- Synology Compose, .env example, .dockerignore and deployment notes.

## Known limits
- Full Vue 3 migration, backend dependency upgrades and security review are **not** completed.
- Container runtime and integration tests are **not** completed.
- No changes have been pushed to GitHub.
