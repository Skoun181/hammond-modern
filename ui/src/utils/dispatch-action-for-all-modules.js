export default function dispatchActionForAllModules(store, modules, actionName) {
  let hasGlobalAction = false

  function visit(modules, namespace = '') {
    for (const [name, definition] of Object.entries(modules)) {
      const prefix = definition.namespaced ? namespace + name + '/' : namespace
      if (definition.actions?.[actionName]) {
        if (prefix) store.dispatch(prefix + actionName)
        else hasGlobalAction = true
      }
      if (definition.modules) visit(definition.modules, prefix)
    }
  }

  visit(modules)
  if (hasGlobalAction) store.dispatch(actionName)
}
