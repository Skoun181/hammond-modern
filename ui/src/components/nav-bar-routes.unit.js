import NavBarRoutes from './nav-bar-routes.vue'

const mountRoutes = (options) => {
  return mount(NavBarRoutes, {
    ...options,
    stubs: { BaseLink: { template: '<a><slot /></a>' }, ...options.stubs },
  })
}

describe('@components/nav-bar-routes', () => {
  it('correctly renders routes with text titles', () => {
    const { element } = mountRoutes({
      props: {
        routes: [
          {
            name: 'aaa',
            title: 'bbb',
          },
        ],
      },
    })
    expect(element.textContent.trim()).toEqual('bbb')
  })

  it('correctly renders routes with function titles', () => {
    const { element } = mountRoutes({
      props: {
        routes: [
          {
            name: 'aaa',
            title: () => 'bbb',
          },
        ],
      },
    })
    expect(element.textContent.trim()).toEqual('bbb')
  })
})
