import { createRouter, createWebHistory } from 'vue-router'
import NoteListView from '../views/NoteListView.vue'
import NoteEditView from '../views/NoteEditView.vue'
import NoteDetailView from '../views/NoteDetailView.vue'
import NotFoundView from '../views/NotFoundView.vue'

const routes = [
  { path: '/', component: NoteListView },
  { path: '/notes/new', component: NoteEditView },
  { path: '/notes/:id', component: NoteDetailView },
  { path: '/notes/:id/edit', component: NoteEditView },
  { path: '/:pathMatch(.*)*', component: NotFoundView },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})
