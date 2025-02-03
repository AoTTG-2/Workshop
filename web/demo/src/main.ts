import {createApp} from 'vue'
import App from './App.vue'
import router from './router'
import './assets/styles.css'
import {auth, initAuth} from './store/auth'
import {setApiBase, setDebugAuth} from "./sdk";

initAuth()
if (auth.userId) {
    setDebugAuth(auth.userId, auth.userRoles)
}

const host = import.meta.env.VITE_WORKSHOP_HOST;
if (host) {
    setApiBase(`${host}/api`);
}

createApp(App).use(router).mount('#app')
