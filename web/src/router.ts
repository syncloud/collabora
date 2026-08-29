import { createRouter, createWebHistory } from 'vue-router'
import Files from './views/Files.vue'
import Editor from './views/Editor.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'files', component: Files },
    { path: '/edit/:id', name: 'editor', component: Editor, props: true },
  ],
})
