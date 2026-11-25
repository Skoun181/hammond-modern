import axios from 'axios'

const userKey = 'auth.currentUser'

function readState(key) {
  try { return JSON.parse(window.localStorage.getItem(key)) }
  catch (_) { return null }
}

function remember(key, value) {
  try { window.localStorage.setItem(key, JSON.stringify(value)) }
  catch (_) { /* Storage is optional. */ }
}

function userDetails(user) {
  if (!user) return null
  const { name, email, role, roleDetail } = user
  return { name, email, role: roleDetail ? roleDetail.long : role }
}

const rememberedUser = userDetails(readState(userKey))
remember(userKey, rememberedUser)
let refreshing
const unauthedClient = axios.create()

function refreshSession() {
  if (!refreshing) {
    refreshing = unauthedClient.post('/api/refresh', {}).then(response => response.data)
      .finally(() => { refreshing = undefined })
  }
  return refreshing
}

axios.interceptors.response.use(response => response, async error => {
  const config = error.config
  if (!config || !error.response || error.response.status !== 401 ||
      !config.url || !config.url.startsWith('/api/') || config.url === '/api/me' || config.sessionRetried)
    throw error
  await refreshSession()
  return axios({ ...config, sessionRetried: true })
})

export async function authenticatedFetch(path, options) {
  let response = await fetch(path, { ...options, credentials: 'same-origin' })
  if (response.status === 401) {
    await refreshSession()
    response = await fetch(path, { ...options, credentials: 'same-origin' })
  }
  return response
}

export const state = {
  currentUser: rememberedUser,
  initialized: readState('system.initialized'),
}

export const mutations = {
  SET_CURRENT_USER(state, value) {
    state.currentUser = userDetails(value)
    remember(userKey, state.currentUser)
  },
  SET_INITIALIZATION_STATUS(state, value) {
    state.initialized = value
    remember('system.initialized', value)
  },
}

export const getters = {
  loggedIn: state => !!state.currentUser,
  isInitialized: state => state.initialized == null || state.initialized.initialized,
}

export const actions = {
  init({ dispatch }) {
    return Promise.all([dispatch('systemInitialized'), dispatch('validate')])
  },

  async logIn({ commit }, { username, password }) {
    const response = await axios.post('/api/login', { email: username, password })
    const user = userDetails(response.data)
    commit('SET_CURRENT_USER', user)
    return user
  },

  async logOut({ commit }) {
    await axios.post('/api/logout', {})
    commit('SET_CURRENT_USER', null)
  },

  async validate({ commit, state }) {
    if (!state.currentUser) return null
    try {
      let user
      try { user = (await axios.get('/api/me')).data }
      catch (error) {
        if (!error.response || error.response.status !== 401) throw error
        user = await refreshSession()
      }
      const details = userDetails(user)
      commit('SET_CURRENT_USER', details)
      return details
    } catch (_) {
      commit('SET_CURRENT_USER', null)
      return null
    }
  },

  async systemInitialized({ commit }) {
    try {
      const { data } = await axios.get('/api/system/status')
      commit('SET_INITIALIZATION_STATUS', data)
      return data
    } catch (_) { return null }
  },
}
