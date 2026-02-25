import 'bootstrap/dist/css/bootstrap.min.css'
import './assets/main.css'
import 'bootstrap/dist/js/bootstrap.bundle.min.js'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { library } from '@fortawesome/fontawesome-svg-core'
import {
  faCopy,
  faCheck,
  faCircleCheck,
  faSpinner,
  faCircleXmark,
  faCircleQuestion,
  faPen,
} from '@fortawesome/free-solid-svg-icons'

import App from './App.vue'
import router from './router'

library.add(faCopy, faCheck, faCircleCheck, faSpinner, faCircleXmark, faCircleQuestion, faPen)

const app = createApp(App)
app.component('FontAwesomeIcon', FontAwesomeIcon)

app.use(createPinia())
app.use(router)

app.mount('#app')
