import Home from './home.vue'

describe('@views/home', () => {
  it('is a valid view', () => {
    expect(Home).toBeAViewComponent()
  })

  it('renders an element', () => {
    const { element } = shallowMountView(Home, {
      ...createComponentMocks({
        store: {
          vehicles: { state: { vehicles: [] }, actions: { fetchVehicles: () => Promise.resolve([]) } },
          users: { state: { me: { currency: 'EUR', vehicles: [] } } },
          utils: { state: { isMobile: false } },
        },
        router: true,
        stubs: { 'b-icon': true, 'b-button': true, 'b-collapse': true,
          'b-dropdown': true, 'b-dropdown-item': true, 'b-notification': true },
      }),
      props: { user: { currency: 'EUR', vehicles: [] } },
    })
    expect(element.textContent).toContain('Dein Dashboard')
  })
})
