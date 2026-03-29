import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import App from './App.vue'
import router from './router'
import { permission } from './directives/permission'

const app = createApp(App)
const pinia = createPinia()
app.use(ElementPlus, {
  locale: zhCn
})
app.use(router)
app.use(pinia)
app.directive('permission', permission)
app.mount('#app')