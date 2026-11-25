import BaseButton from './_base-button.vue'

describe('@components/_base-button', () => {
  it('renders its content', () => {
    const slotContent = '<span>foo</span>'
    const { element } = shallowMount(BaseButton, {
      mocks: { $style: {} },
      stubs: { 'b-button': { template: '<button><slot /></button>' } },
      slots: {
        default: slotContent,
      },
    })
    expect(element.innerHTML).toContain(slotContent)
  })
})
