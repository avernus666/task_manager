# Upgrade notes

This version contains a deployment and usability pass on Task Manager Rework.

## Deployment and infrastructure
- API Docker image includes SQL migrations and runs pending migrations before the API starts (`AUTO_MIGRATE=true` by default).
- Docker Compose no longer needs a separate migration service; the API waits for PostgreSQL health first.
- Frontend Docker image now builds Vite assets and serves the production build with Nginx instead of running the Vite development server.
- Frontend image accepts `VITE_API_URL` as a build argument so the API URL is baked into the static bundle correctly.
- Added SPA fallback and cache headers in `frontend/nginx.conf`.
- Render blueprint uses the same migration-on-start behavior.

## App improvements
- Task list can be filtered by status, priority, and project at once.
- Task creation supports a deadline.
- Task detail drawer supports deleting a task with confirmation.
- Dashboard date is dynamic instead of hard-coded.
- API connection badge reflects workspace-load success/failure.
- Dashboard shows a count of overdue unfinished tasks.
- API rejects whitespace-only usernames, project names, and task titles; assignees must be active users.
- Initial schema migrations are tolerant of the existing users/projects tables and indexes, which helps recover a database where those first tables were created before the migration history was initialized.

## Build verification
YAML syntax for `compose.yaml` and `render.yaml` was checked. A full Go/TypeScript build was not possible in this environment because dependency registries were unreachable; run the repository CI workflow before production deployment.


## Final polish
- Fixed the test-email action so it saves the currently entered notification address and preferences before requesting delivery.
- Added the existing SVG app icon to the PWA manifest for installable-browser metadata.
- Added visible keyboard focus, text selection styling, and touch-device hover safeguards.
- Build verification remains pending because this environment cannot reach public dependency registries; run the repository CI workflow before production deployment.
