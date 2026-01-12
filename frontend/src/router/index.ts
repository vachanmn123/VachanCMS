import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'login',
      component: () => import('@/views/Login.vue'),
    },
    {
      path: '/repos',
      name: 'repos',
      component: () => import('@/views/RepoSelect.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/dashboard/:owner/:repo',
      component: () => import('@/views/DashboardLayout.vue'),
      props: true,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'content-types',
          component: () => import('@/views/ContentTypesView.vue'),
        },
        {
          path: 'media',
          name: 'media',
          component: () => import('@/views/MediaView.vue'),
        },
        {
          path: ':ctSlug',
          name: 'content-values',
          component: () => import('@/views/ContentValuesView.vue'),
          props: true,
        },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    await authStore.checkAuth()
    if (!authStore.isAuthenticated) {
      return '/'
    }
  }
})

export default router
