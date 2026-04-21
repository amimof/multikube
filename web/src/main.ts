import { createApp } from 'vue'
import { createPinia } from 'pinia'
import "element-plus/dist/index.css";
import "element-plus/theme-chalk/dark/css-vars.css";
import "./styles/main.css";
// import 'element-plus/theme-chalk/dark/css-vars.css'
// import '@/styles/element/theme.scss'
// import '@/styles/element/dark-overrides.css'
import ElementPlus from 'element-plus'


import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus)

app.mount('#app')
