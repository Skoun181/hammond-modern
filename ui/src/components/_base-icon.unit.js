import BaseIcon from './_base-icon.vue'

describe('@components/_base-icon', () => {
  it('renders a font-awesome icon', () => {
    const { element } = mount(BaseIcon, {
      props: {
        name: 'sync',
      },
    })

    expect(element.tagName).toEqual('svg')
    expect(element.classList).toContain('svg-inline--fa', 'fa-sync', 'fa-w-16')
  })

  it('renders a custom icon', () => {
    const { element } = shallowMount(BaseIcon, {
      ...createComponentMocks({
        style: {
          iconCustomSomeIcon: 'generated-class-name',
        },
      }),
      props: {
        source: 'custom',
        name: 'some-icon',
      },
    })

    expect(element.className).toEqual('generated-class-name')
  })
})
