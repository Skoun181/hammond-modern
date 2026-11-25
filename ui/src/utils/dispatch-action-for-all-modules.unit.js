import { createStore } from 'vuex'
import dispatchActionForAllModules from './dispatch-action-for-all-modules'

describe('@utils/dispatch-action-for-all-modules', () => {
  for (const namespaced of [false, true]) {
    for (const nested of [false, true]) {
      it(`dispatches ${namespaced ? 'namespaced' : 'shared'} actions from ${nested ? 'nested' : 'top-level'} modules`, () => {
        const one = { namespaced, actions: { start: vi.fn(), ignored: vi.fn() } }
        const two = { namespaced, actions: { start: vi.fn(), ignored: vi.fn() } }
        const modules = nested ? { one: { ...one, modules: { two } } } : { one, two }
        dispatchActionForAllModules(createStore({ modules }), modules, 'start')
        for (const module of [one, two]) {
          expect(module.actions.start).toHaveBeenCalledTimes(1)
          expect(module.actions.ignored).not.toHaveBeenCalled()
        }
      })
    }
  }
})
