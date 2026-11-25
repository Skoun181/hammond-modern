import * as usersModule from './users'

describe('@state/modules/users', () => {
  it('exports a valid Vuex module', () => {
    expect(usersModule).toBeAVuexModule()
  })

  describe('in a store when logged in', () => {
    let store
    beforeEach(() => {
      store = createModuleStore(usersModule, {
        currentUser: validUserExample,
      })
    })

    it('actions.fetchUser returns the current user without fetching it again', () => {
      expect.assertions(2)

      const axios = require('axios')
      const originalAxiosGet = axios.get
      axios.get = jest.fn()

      return store.dispatch('fetchUser', { username: 'admin' }).then((user) => {
        expect(user).toEqual(validUserExample)
        expect(axios.get).not.toHaveBeenCalled()
        axios.get = originalAxiosGet
      })
    })

    it('returns no cached user when the username is unknown', async () => {
      expect(await store.dispatch('fetchUser', { username: 'missing' })).toBeUndefined()
    })
  })
})

const validUserExample = { id: 1, username: 'admin', name: 'Test user' }
