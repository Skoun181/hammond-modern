import Login from './login.vue'

describe('@views/login', () => {
  it('is a valid view', () => {
    expect(Login).toBeAViewComponent()
  })

  it('redirects to home after successful login', async () => {
    const routerPush = vi.fn()
    const { vm } = mountLogin(routerPush)
    vm.username = 'correctUsername'
    vm.password = 'correctPassword'
    await vm.tryToLogIn()
    expect(vm.authError).toBeNull()
    expect(routerPush).toHaveBeenCalledWith({ name: 'home' })
  })

  it('returns to the requested route after successful login', async () => {
    const routerPush = vi.fn()
    const redirectFrom = '/profile?someQuery'
    const { vm } = mountLogin(routerPush, { redirectFrom })
    vm.username = 'correctUsername'
    vm.password = 'correctPassword'
    await vm.tryToLogIn()
    expect(vm.authError).toBeNull()
    expect(routerPush).toHaveBeenCalledWith(redirectFrom)
  })

  it('displays an error after failed login', async () => {
    const { vm } = mountLogin(vi.fn())
    await vm.tryToLogIn()
    await vm.$nextTick()
    expect(vm.authError).toBeTruthy()
    expect(vm.$el.textContent).toContain('error')
  })
})

function mountLogin(routerPush, query = {}) {
  return shallowMountView(Login, {
    ...createComponentMocks({
      mocks: { $router: { push: routerPush }, $route: { query } },
      stubs: { 'b-field': true, 'b-input': true, 'b-button': true },
      store: {
        users: { actions: { forceMe: () => Promise.resolve({}) } },
        auth: {
          actions: {
            logIn(_, { username, password }) {
              if (
                username === 'correctUsername' &&
                password === 'correctPassword'
              ) {
                return Promise.resolve('testToken')
              } else {
                return Promise.reject(new Error('testError'))
              }
            },
          },
        },
      },
    }),
  })
}
