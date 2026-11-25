# Frontend development

Use Node.js 24 and npm. The Vite dev server listens on port 8080 and forwards API requests to the Go server on port 3000 by default.

```bash
npm ci
npm run dev
```

Set `API_BASE_URL=http://localhost:3000` to choose another backend for local development. Run that backend with separate development data. Vite keeps this server-side configuration out of the browser bundle.

`npm run build` writes the production assets to `dist/`. The root Dockerfile builds both the frontend and the Go server into the Hammond image.

See [tests](tests.md) for the Vitest and browser checks.
