import store from '@state/store'

export default [
  {
    path: '/',
    name: 'home',
    meta: {
      authRequired: true,
      tmp: {},
      beforeResolve(routeTo, routeFrom, next) {
        store
          // Try to fetch the user's information by their username
          .dispatch('vehicles/fetchVehicles')
          .then((vehicles) => {
            // Add the user to `meta.tmp`, so that it can
            // be provided as a prop.
            routeTo.meta.tmp.vehicles = vehicles

            store
              // Try to fetch the user's information by their username
              .dispatch('users/me')
              .then((me) => {
                next()
              })
            // Continue to the route.
          })
          .catch((ex) => {
            // If a user with the provided username could not be
            // found, redirect to the 404 page.
            console.log(ex)
            next({ name: '404', query: { resource: 'User' } })
          })
      },
    },
    component: () => import('@views/home.vue'),
    props: (route) => ({ user: store.state.auth.currentUser || {} }),
  },
  { path: '/recurring-costs', name: 'recurring-costs', component: () => import('@views/recurringCosts.vue'), meta: { authRequired: true } },
  { path: '/home-assistant', name: 'home-assistant', component: () => import('@views/homeAssistant.vue'), meta: { authRequired: true } },
  { path: '/maintenance', name: 'maintenance', component: () => import('@views/maintenance.vue'), meta: { authRequired: true } },
  { path: '/vehicles', name: 'vehicles', component: () => import('@views/vehicles.vue'), meta: { authRequired: true } },
  { path: '/tires', name: 'tires', component: () => import('@views/tires.vue'), meta: { authRequired: true } },
  { path: '/drivers', name: 'drivers', component: () => import('@views/drivers.vue'), meta: { authRequired: true } },
  {
    path: '/initialize',
    name: 'initialize',
    component: () => import('@views/initialize.vue'),
    meta: {
      async beforeResolve(routeTo, routeFrom, next) {
        const system = await store.dispatch('auth/systemInitialized')
        if (!system || system.initialized) return next({ name: 'login' })
        return next()
      },
    },
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@views/login.vue'),
    meta: {
      async beforeResolve(routeTo, routeFrom, next) {
        const system = await store.dispatch('auth/systemInitialized')
        if (system && !system.initialized) return next({ name: 'initialize' })
        if (store.getters['auth/loggedIn']) {
          const user = await store.dispatch('auth/validate')
          if (user) return next({ name: 'home' })
        }
        return next()
      },
    },
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('@views/profile.vue'),
    meta: {
      authRequired: true,
      roles: ['ADMIN'],
    },
    props: (route) => ({ user: store.state.auth.currentUser || {} }),
  },
  {
    path: '/admin/settings',
    name: 'site-settings',
    component: () => import('@/src/router/views/siteSettings.vue'),
    meta: {
      authRequired: true,
      roles: ['ADMIN'],
    },
    props: (route) => ({
      user: store.state.auth.currentUser || {},
      settings: store.state.utils.settings || {},
    }),
  },
  {
    path: '/admin/users',
    name: 'users',
    component: () => import('@/src/router/views/users.vue'),
    meta: {
      authRequired: true,
      roles: ['ADMIN'],
    },
    props: (route) => ({
      user: store.state.auth.currentUser || {},
      settings: store.state.utils.settings || {},
    }),
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('@/src/router/views/settings.vue'),
    meta: {
      authRequired: true,
    },
    props: (route) => ({
      user: store.state.auth.currentUser || {},
      me: store.state.users.me || {},
    }),
  },
  {
    path: '/profile/:username',
    name: 'username-profile',
    component: () => import('@views/profile.vue'),
    meta: {
      authRequired: true,
      // HACK: In order to share data between the `beforeResolve` hook
      // and the `props` function, we must create an object for temporary
      // data only used during route resolution.
      tmp: {},
      beforeResolve(routeTo, routeFrom, next) {
        store
          // Try to fetch the user's information by their username
          .dispatch('users/fetchUser', { username: routeTo.params.username })
          .then((user) => {
            // Add the user to `meta.tmp`, so that it can
            // be provided as a prop.
            routeTo.meta.tmp.user = user
            // Continue to the route.
            next()
          })
          .catch(() => {
            // If a user with the provided username could not be
            // found, redirect to the 404 page.
            next({ name: '404', query: { resource: 'User' } })
          })
      },
    },
    // Set the user from the route params, once it's set in the
    // beforeResolve route guard.
    props: (route) => ({ user: route.meta.tmp.user }),
  },

  {
    path: '/vehicles/create',
    name: 'vehicle-create',
    component: () => import('@views/createVehicle.vue'),
    meta: {
      authRequired: true,
      // HACK: In order to share data between the `beforeResolve` hook
      // and the `props` function, we must create an object for temporary
      // data only used during route resolution.
    },
  },
  {
    path: '/vehicles/:vehicleId',
    name: 'vehicle-detail',
    component: () => import('@views/vehicle.vue'),
    meta: {
      authRequired: true,
      // HACK: In order to share data between the `beforeResolve` hook
      // and the `props` function, we must create an object for temporary
      // data only used during route resolution.
      tmp: {},
      beforeResolve(routeTo, routeFrom, next) {
        store
          // Try to fetch the user's information by their username
          .dispatch('vehicles/fetchVehicleById', {
            vehicleId: routeTo.params.vehicleId,
          })
          .then((vehicle) => {
            // Add the user to `meta.tmp`, so that it can
            // be provided as a prop.
            routeTo.meta.tmp.vehicle = vehicle
            // Continue to the route.
            next()
          })
          .catch(() => {
            // If a user with the provided username could not be
            // found, redirect to the 404 page.
            next({ name: '404', query: { resource: 'User' } })
          })
      },
    },
    // Set the user from the route params, once it's set in the
    // beforeResolve route guard.
    props: (route) => ({ vehicle: route.meta.tmp.vehicle }),
  },
  {
    path: '/vehicles/:vehicleId/edit',
    name: 'vehicle-edit',
    component: () => import('@views/createVehicle.vue'),
    meta: {
      authRequired: true,
      // HACK: In order to share data between the `beforeResolve` hook
      // and the `props` function, we must create an object for temporary
      // data only used during route resolution.
      tmp: {},
      beforeResolve(routeTo, routeFrom, next) {
        store
          // Try to fetch the user's information by their username
          .dispatch('vehicles/fetchVehicleById', {
            vehicleId: routeTo.params.vehicleId,
          })
          .then((vehicle) => {
            // Add the user to `meta.tmp`, so that it can
            // be provided as a prop.
            routeTo.meta.tmp.vehicle = vehicle
            // Continue to the route.
            next()
          })
          .catch(() => {
            // If a user with the provided username could not be
            // found, redirect to the 404 page.
            next({ name: '404', query: { resource: 'User' } })
          })
      },
    },
    // Set the user from the route params, once it's set in the
    // beforeResolve route guard.
    props: (route) => ({ vehicle: route.meta.tmp.vehicle }),
  },
  {
    path: '/vehicles/:vehicleId/fillup',
    name: 'vehicle-create-fillup',
    component: () => import('@views/createFillup.vue'),
    meta: {
      authRequired: true,
      // HACK: In order to share data between the `beforeResolve` hook
      // and the `props` function, we must create an object for temporary
      // data only used during route resolution.
      tmp: {},
      beforeResolve(routeTo, routeFrom, next) {
        store
          // Try to fetch the user's information by their username
          .dispatch('vehicles/fetchVehicleById', {
            vehicleId: routeTo.params.vehicleId,
          })
          .then((vehicle) => {
            // Add the user to `meta.tmp`, so that it can
            // be provided as a prop.
            routeTo.meta.tmp.vehicle = vehicle
            // Continue to the route.
            next()
          })
          .catch(() => {
            // If a user with the provided username could not be
            // found, redirect to the 404 page.
            next({ name: '404', query: { resource: 'User' } })
          })
      },
    },
    // Set the user from the route params, once it's set in the
    // beforeResolve route guard.
    props: (route) => ({ vehicle: route.meta.tmp.vehicle }),
  },
  {
    path: '/vehicles/:vehicleId/fillup/:fillupId/edit',
    name: 'vehicle-edit-fillup',
    component: () => import('@views/createFillup.vue'),
    meta: {
      authRequired: true,

      tmp: {},
      beforeResolve(routeTo, routeFrom, next) {
        store

          .dispatch('vehicles/fetchVehicleById', {
            vehicleId: routeTo.params.vehicleId,
          })
          .then((vehicle) => {
            routeTo.meta.tmp.vehicle = vehicle
            store

              .dispatch('vehicles/fetchFillupById', {
                vehicleId: routeTo.params.vehicleId,
                fillupId: routeTo.params.fillupId,
              })
              .then((fillup) => {
                routeTo.meta.tmp.fillup = fillup

                next()
              })
          })
          .catch(() => {
            next({ name: '404', query: { resource: 'User' } })
          })
      },
    },
    // Set the user from the route params, once it's set in the
    // beforeResolve route guard.
    props: (route) => ({ vehicle: route.meta.tmp.vehicle, fillup: route.meta.tmp.fillup }),
  },
  {
    path: '/vehicles/:vehicleId/expense',
    name: 'vehicle-create-expense',
    component: () => import('@views/createExpense.vue'),
    meta: {
      authRequired: true,
      // HACK: In order to share data between the `beforeResolve` hook
      // and the `props` function, we must create an object for temporary
      // data only used during route resolution.
      tmp: {},
      beforeResolve(routeTo, routeFrom, next) {
        store
          // Try to fetch the user's information by their username
          .dispatch('vehicles/fetchVehicleById', {
            vehicleId: routeTo.params.vehicleId,
          })
          .then((vehicle) => {
            // Add the user to `meta.tmp`, so that it can
            // be provided as a prop.
            routeTo.meta.tmp.vehicle = vehicle
            // Continue to the route.
            next()
          })
          .catch(() => {
            // If a user with the provided username could not be
            // found, redirect to the 404 page.
            next({ name: '404', query: { resource: 'User' } })
          })
      },
    },
    // Set the user from the route params, once it's set in the
    // beforeResolve route guard.
    props: (route) => ({ vehicle: route.meta.tmp.vehicle }),
  },
  {
    path: '/vehicles/:vehicleId/expense/:expenseId/edit',
    name: 'vehicle-edit-expense',
    component: () => import('@views/createExpense.vue'),
    meta: {
      authRequired: true,

      tmp: {},
      beforeResolve(routeTo, routeFrom, next) {
        store

          .dispatch('vehicles/fetchVehicleById', {
            vehicleId: routeTo.params.vehicleId,
          })
          .then((vehicle) => {
            routeTo.meta.tmp.vehicle = vehicle
            store

              .dispatch('vehicles/fetchExpenseById', {
                vehicleId: routeTo.params.vehicleId,
                expenseId: routeTo.params.expenseId,
              })
              .then((expense) => {
                routeTo.meta.tmp.expense = expense

                next()
              })
          })
          .catch(() => {
            next({ name: '404', query: { resource: 'User' } })
          })
      },
    },
    // Set the user from the route params, once it's set in the
    // beforeResolve route guard.
    props: (route) => ({ vehicle: route.meta.tmp.vehicle, expense: route.meta.tmp.expense }),
  },
  {
    path: '/admin/backup',
    name: 'backup',
    component: () => import('@views/backup.vue'),
    meta: { authRequired: true, roles: ['ADMIN'] },
  },
  {
    path: '/import',
    name: 'import',
    component: () => import('@views/import.vue'),
    meta: {
      authRequired: true,
    },
    props: (route) => ({ user: store.state.auth.currentUser || {} }),
  },
  {
    path: '/import/hammond',
    name: 'import-hammond',
    component: () => import('@views/importHammond.vue'),
    meta: { authRequired: true },
  },
  {
    path: '/import/fuelly',
    name: 'import-fuelly',
    component: () => import('@views/import-fuelly.vue'),
    meta: {
      authRequired: true,
    },
    props: (route) => ({ user: store.state.auth.currentUser || {} }),
  },
  {
    path: '/logout',
    name: 'logout',
    component: () => import('@views/_loading.vue'),
    meta: {
      authRequired: true,
      beforeResolve(routeTo, routeFrom, next) {
        store.dispatch('auth/logOut').then((data) => {
          const authRequiredOnPreviousRoute = routeFrom.matched.some((route) => route.meta.authRequired)
          // Navigate back to previous page, or home as a fallback
          next(authRequiredOnPreviousRoute || !routeFrom.name ? { name: 'login' } : routeFrom.fullPath)
        })
      },
    },
  },
  {
    path: '/404',
    name: '404',
    component: () => import('@views/_404.vue'),
    // Allows props to be passed to the 404 page through route
    // params, such as `resource` to define what wasn't found.
    props: (route) => ({ resource: route.query.resource }),
  },
  // Redirect any unmatched routes to the 404 page. This may
  // require some server configuration to work in production:
  // https://router.vuejs.org/en/essentials/history-mode.html#example-server-configurations
  {
    path: '/:pathMatch(.*)*',
    redirect: { name: '404' },
  },
]
