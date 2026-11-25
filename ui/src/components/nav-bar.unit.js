import NavBar from './nav-bar.vue'

describe('@components/nav-bar', () => {
  it(`shows administrative actions only to an administrator`, () => {
    const { vm } = shallowMount(
      NavBar,
      createComponentMocks({
        router: true,
        mocks: { $route: { name: 'home' } },
        store: {
          auth: {
            state: {
              currentUser: {
                name: 'My Name',
                role: 'ADMIN',
              },
            },
            getters: {
              loggedIn: () => true,
            },
          },
        },
      })
    )

    expect(vm.links.map((link) => link.name)).toContain('site-settings')
    expect(vm.links.map((link) => link.name)).toContain('backup')
  })
})
