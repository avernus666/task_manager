# Production deployment

## Target architecture

- **Vercel** serves `frontend/` as the static React/Vite application.
- **Render** runs the Go API from the repository Dockerfile.
- **Neon** hosts PostgreSQL. The application must use Neon's pooled connection string with TLS enabled.
- SMTP is configured as Render environment variables; credentials are never sent to the browser.

## First deployment checklist

1. Create a Neon database/project and copy its pooled PostgreSQL connection string. Keep `sslmode=require` (or the provider's equivalent TLS settings).
2. Deploy the API service on Render from this repository using `render.yaml`. Set `DATABASE_URL` to the Neon connection string, `CORS_ALLOWED_ORIGINS` to the exact Vercel origin, and set SMTP variables if email delivery is required.
3. The API Docker image contains the SQL migrations and applies pending migrations on startup (`AUTO_MIGRATE=true` by default). Confirm `/health/ready` returns `200` after the first start. For maintenance windows where migrations must be controlled separately, set `AUTO_MIGRATE=false` and run a trusted migration runner against the same database.
4. On Vercel, import the same repository and set **Root Directory** to `frontend`. Set `VITE_API_URL` to the Render API origin, without a trailing slash. Deploy.
5. Return to Render and set `CORS_ALLOWED_ORIGINS` to the final Vercel origin, for example `https://your-project.vercel.app`, then redeploy.
6. Verify `https://YOUR-API/health/live` and `/health/ready`, then register a test account and create a project.
7. Configure SMTP and use the in-app test email action.

## Important deployment notes

- The `render.yaml` file intentionally marks database, CORS and SMTP secrets as unsynced values. Enter them in Render's dashboard; never commit real credentials.
- Render and Vercel need access to the Git repository. This repository must be pushed to a GitHub account that both services can read.
- Database migrations are run by the Docker entrypoint before the API starts (`AUTO_MIGRATE=true`). The API binary itself does not run migrations when launched outside the Docker image.
- The free service tier can sleep or have cold starts. Availability and quotas depend on the selected plans.
- Browser notifications require permission and an installed/open app context. True push delivery while the app is closed requires Web Push/VAPID and persisted push subscriptions; this codebase does not claim that capability is already enabled.
- Before public launch, run dependency scanning, integration tests, a backup/restore rehearsal and a production smoke test. The in-memory rate limiter is per instance and should be replaced by a shared limiter for multi-instance production.
