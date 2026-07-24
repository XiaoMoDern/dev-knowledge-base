import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import NoteListView from '../views/NoteListView.vue'
import NoteEditView from '../views/NoteEditView.vue'
import NoteDetailView from '../views/NoteDetailView.vue'
import NotFoundView from '../views/NotFoundView.vue'

const routes = [
  {
    path: '/',
    component: MainLayout,
    children: [
      { path: '', component: NoteListView },
      { path: 'notes/new', component: NoteEditView, meta: { hideSidebar: true } },
      { path: 'notes/:id', component: NoteDetailView, meta: { hideSidebar: true } },
      { path: 'notes/:id/edit', component: NoteEditView, meta: { hideSidebar: true } },
    ],
  },
  { path: '/:pathMatch(.*)*', component: NotFoundView },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})
