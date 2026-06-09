import type { RouteRecordRaw } from 'vue-router'
import Layout from '@/components/Layout/index.vue'

const platformRoutes: RouteRecordRaw[] = [
  {
    path: '/platform',
    component: Layout,
    redirect: '/platform/system/user',
    meta: { title: '平台管理系统' },
    children: [
      {
        path: 'system',
        name: 'PlatformSystem',
        meta: { title: '系统管理', icon: 'Setting' },
        redirect: '/platform/system/user',
        children: [
          {
            path: 'user',
            name: 'PlatformUser',
            component: () => import('@/views/platform/system/user/index.vue'),
            meta: { title: '用户管理', icon: 'User', permissionCode: 'platform:user:list' },
          },
          {
            path: 'role',
            name: 'PlatformRole',
            component: () => import('@/views/platform/system/role/index.vue'),
            meta: { title: '角色管理', icon: 'Key', permissionCode: 'platform:role:list' },
          },
          {
            path: 'dept',
            name: 'PlatformDept',
            component: () => import('@/views/platform/system/dept/index.vue'),
            meta: { title: '部门管理', icon: 'OfficeBuilding', permissionCode: 'platform:dept:list' },
          },
        ],
      },
      {
        path: 'shop',
        name: 'PlatformShop',
        component: () => import('@/views/platform/shop/index.vue'),
        meta: { title: '店铺管理', icon: 'Shop', permissionCode: 'platform:shop:list' },
      },
      {
        path: 'product',
        name: 'PlatformProduct',
        meta: { title: '商品管理', icon: 'Goods' },
        redirect: '/platform/product/category',
        children: [
          {
            path: 'category',
            name: 'PlatformProductCategory',
            component: () => import('@/views/platform/product/category/index.vue'),
            meta: { title: '商品分类', icon: 'List', permissionCode: 'platform:product:category:list' },
          },
          {
            path: 'list',
            name: 'PlatformProductList',
            component: () => import('@/views/platform/product/list/index.vue'),
            meta: { title: '商品列表', icon: 'Goods', permissionCode: 'platform:product:list' },
          },
        ],
      },
      {
        path: 'finance',
        name: 'PlatformFinance',
        meta: { title: '财务管理', icon: 'Money' },
        redirect: '/platform/finance/category',
        children: [
          {
            path: 'category',
            name: 'PlatformFinanceCategory',
            component: () => import('@/views/platform/finance/category/index.vue'),
            meta: { title: '收支分类', icon: 'List', permissionCode: 'platform:finance:category:list' },
          },
        ],
      },
    ],
  },
]

export default platformRoutes
