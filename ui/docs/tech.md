# Frontend stack

- Vue 3 and Buefy 3 render the application. Vue Router handles routes and Vuex holds shared state.
- Vite serves the frontend during development and builds the production assets. Its settings and import aliases are in `vite.config.mjs`.
- Chart.js 4 renders the vehicle consumption chart.
- Vitest runs component and state tests. Playwright checks the UI against the real Go backend. [Test commands](tests.md)
