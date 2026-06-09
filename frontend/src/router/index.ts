import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import platformRoutes from './modules/platform'
import shopRoutes from './modules/shop'

const staticRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/platform/login',
  },
  {
    path: '/platform/login',
    name: 'PlatformLogin',
    component: () => import('@/views/platform/login/index.vue'),
    meta: { title: '平台登录' },
  },
  {
    path: '/shop/login',
    name: 'ShopLogin',
    component: () => import('@/views/shop/login/index.vue'),
    meta: { title: '店铺登录' },
  },
  ...platformRoutes,
  ...shopRoutes,
]

const router = createRouter({
  history: createWebHistory(),
  routes: staticRoutes,
})

export default router
