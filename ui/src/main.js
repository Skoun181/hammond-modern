import { applySavedTheme } from '@utils/theme-modern'
import { createApp } from 'vue'
import Buefy from 'buefy'
import router from '@router'
import store, { initializeStore } from '@state/store'
import { library } from '@fortawesome/fontawesome-svg-core'
import {
  faCheck,
  faTimes,
  faArrowUp,
  faAngleLeft,
  faAngleRight,
  faCalendar,
  faEdit,
  faAngleDown,
  faAngleUp,
  faUpload,
  faExclamationCircle,
  faDownload,
  faEye,
  faEyeSlash,
  faTrash,
  faShare,
  faUserFriends,
  faTimesCircle,
} from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import App from './app.vue'
import appConfig from './app.config.json'

// Globally register all `_base`-prefixed components
import { registerBaseComponents } from '@components/_globals'

import 'buefy/dist/css/buefy.css'
import 'nprogress/nprogress.css'
import './design/modern.css'

applySavedTheme()

const app = createApp(App)
app.component('vue-fontawesome', FontAwesomeIcon)
registerBaseComponents(app)
library.add(
  faCheck,
  faTimes,
  faArrowUp,
  faAngleLeft,
  faAngleRight,
  faCalendar,
  faEdit,
  faAngleDown,
  faAngleUp,
  faUpload,
  faExclamationCircle,
  faDownload,
  faEye,
  faEyeSlash,
  faTrash,
  faShare,
  faUserFriends,
  faTimesCircle
)
app.use(Buefy, {
  defaultIconComponent: 'vue-fontawesome',
  defaultIconPack: 'fas',
})

function updateMetadata() {
  const page = this.$options.page
  if (!page) return
  const metadata = typeof page === 'function' ? page.call(this) : page
  document.title = metadata.title ? `${metadata.title} | ${appConfig.title}` : appConfig.title
  const description = metadata.meta?.find(entry => entry.name === 'description')
  document.querySelector('meta[name="description"]')
    ?.setAttribute('content', description?.content || appConfig.description)
}

app.mixin({ mounted: updateMetadata, updated: updateMetadata })
app.use(store)
initializeStore()
app.use(router)
router.isReady().then(() => app.mount('#app'))
