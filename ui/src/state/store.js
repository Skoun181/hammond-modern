import { createStore } from 'vuex'
import dispatchActionForAllModules from '@utils/dispatch-action-for-all-modules'
import modules from './modules'

const store = createStore({ modules, strict: import.meta.env.DEV })

export function initializeStore() {
  dispatchActionForAllModules(store, modules, 'init')
}

export default store
