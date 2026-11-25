import axios from 'axios'
import * as auth from './auth'

describe('@state/modules/auth', () => {
  let store
  const user = { name: 'Alex', email: 'alex@example.test', role: 'USER' }

  beforeEach(() => {
    window.localStorage.clear()
    store = createModuleStore(auth)
  })

  afterEach(() => jest.restoreAllMocks())

  it('uses the cookie-based login API and stores only account details', async () => {
    const post = jest.spyOn(axios, 'post').mockResolvedValue({
      data: { ...user, token: 'must-not-be-stored', refreshToken: 'must-not-be-stored' },
    })
    expect(await store.dispatch('logIn', { username: user.email, password: 'a password' })).toEqual(user)
    expect(post).toHaveBeenCalledWith('/api/login', { email: user.email, password: 'a password' })
    expect(JSON.parse(window.localStorage.getItem('auth.currentUser'))).toEqual(user)
  })

  it('validates via the current-account route', async () => {
    store.commit('SET_CURRENT_USER', user)
    const get = jest.spyOn(axios, 'get').mockResolvedValue({
      data: { ...user, role: 1, roleDetail: { long: 'USER' } },
    })
    expect(await store.dispatch('validate')).toEqual(user)
    expect(get).toHaveBeenCalledWith('/api/me')
  })

  it('revokes the cookie session on logout', async () => {
    store.commit('SET_CURRENT_USER', user)
    const post = jest.spyOn(axios, 'post').mockResolvedValue({})
    await store.dispatch('logOut')
    expect(post).toHaveBeenCalledWith('/api/logout', {})
    expect(store.state.currentUser).toBeNull()
  })
})
