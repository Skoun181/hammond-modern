# Frontend tests

Run `npm ci`, then use `npm run test:unit` for the Vue component and state tests. Use `npm run test:unit:watch` to rerun them while editing.

The Playwright check starts the real Go server on port 3000 with a temporary database and asset directory. It covers setup, login, vehicle creation, the consumption chart, and ZIP export.

```bash
npm run build
go -C ../server build -o /tmp/hammond-e2e .
npx playwright install chromium
HAMMOND_TEST_BINARY=/tmp/hammond-e2e npm run test:e2e
```

For an existing Chromium installation, set `HAMMOND_TEST_CHROMIUM` to its executable path.
