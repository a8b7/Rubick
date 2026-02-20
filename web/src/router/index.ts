import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/components/layout/AppLayout.vue'),
    meta: { title: 'Rubick' },
  },
  {
    path: '/hosts',
    redirect: '/',
  },
  {
    path: '/containers',
    redirect: '/',
  },
  {
    path: '/images',
    redirect: '/',
  },
  {
    path: '/compose',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫 - 设置页面标题
router.beforeEach((to, _from, next) => {
  document.title = `${to.meta.title || 'Rubick'} - Rubick`
  next()
})

export default router
