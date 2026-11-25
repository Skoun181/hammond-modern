import { config, enableAutoUnmount, mount as mountVue, shallowMount as shallowMountVue } from '@vue/test-utils'
import { createStore } from 'vuex'
import { cloneDeep } from 'lodash'
import { registerBaseComponents } from '@components/_globals'

config.global.plugins = [registerBaseComponents]
enableAutoUnmount(afterEach)

function mountOptions({ mocks, stubs, global, ...options } = {}) {
  return {
    ...options,
    global: {
      ...global,
      mocks: { ...global?.mocks, ...mocks },
      stubs: { ...global?.stubs, ...stubs },
    },
  }
}

globalThis.mount = (component, options) => mountVue(component, mountOptions(options))
globalThis.shallowMount = (component, options) => shallowMountVue(component, mountOptions(options))

globalThis.shallowMountView = (component, options = {}) => {
  return globalThis.shallowMount(component, {
    ...options,
    stubs: { Layout: { template: '<main><slot /></main>' }, ...options.stubs },
  })
}

globalThis.createComponentMocks = ({ store, router, style, mocks, stubs } = {}) => ({
  global: {
    plugins: store ? [createStore({
      modules: Object.fromEntries(
        Object.entries(store).map(([name, definition]) => [name, { namespaced: true, ...definition }])
      ),
    })] : [],
  },
  mocks: { ...mocks, ...(style && { $style: style }) },
  stubs: {
    ...(router && { RouterLink: true, RouterView: true }),
    ...stubs,
  },
})

globalThis.createModuleStore = (module, { currentUser } = {}) => createStore({
  ...cloneDeep({ ...module }),
  modules: { auth: { namespaced: true, state: { currentUser } } },
  strict: true,
})
