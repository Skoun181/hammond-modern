import * as auth from './auth'
import * as users from './users'
import * as utils from './utils'
import * as vehicles from './vehicles'

const modules = { auth, users, utils, vehicles }

export default Object.fromEntries(
  Object.entries(modules).map(([name, definition]) => [
    name,
    { ...definition, namespaced: true },
  ])
)
