const components = import.meta.glob('./_base-*.vue', { eager: true, import: 'default' })

export function registerBaseComponents(app) {
  for (const [path, component] of Object.entries(components)) {
    const name = path.slice(3, -4).split('-')
      .map(part => part.charAt(0).toUpperCase() + part.slice(1)).join('')
    app.component(name, component)
  }
}
