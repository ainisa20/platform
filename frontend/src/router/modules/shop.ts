import type { RouteRecordRaw } from 'vue-router'
import Layout from '@/components/Layout/index.vue'

const shopRoutes: RouteRecordRaw[] = [
  {
    path: '/shop',
    component: Layout,
    redirect: '/shop/system/user',
    meta: { title: '店铺管理系统' },
    children: [
      {
        path: 'system',
        name: 'ShopSystem',
        meta: { title: '系统管理', icon: 'Setting' },
        redirect: '/shop/system/user',
        children: [
          {
            path: 'user',
            name: 'ShopUser',
            component: () => import('@/views/shop/system/user/index.vue'),
            meta: { title: '用户管理', icon: 'User', permissionCode: 'shop:user:list' },
          },
          {
            path: 'role',
            name: 'ShopRole',
            component: () => import('@/views/shop/system/role/index.vue'),
            meta: { title: '角色管理', icon: 'Key', permissionCode: 'shop:role:list' },
          },
          {
            path: 'dept',
            name: 'ShopDept',
            component: () => import('@/views/shop/system/dept/index.vue'),
            meta: { title: '部门管理', icon: 'OfficeBuilding', permissionCode: 'shop:dept:list' },
          },
        ],
      },
      {
        path: 'product',
        name: 'ShopProduct',
        meta: { title: '商品管理', icon: 'Goods' },
        redirect: '/shop/product/list',
        children: [
          {
            path: 'list',
            name: 'ShopProductList',
            component: () => import('@/views/shop/product/index.vue'),
            meta: { title: '选品管理', icon: 'List', permissionCode: 'shop:product:list' },
          },
        ],
      },
      {
        path: 'customer',
        name: 'ShopCustomer',
        component: () => import('@/views/shop/customer/index.vue'),
        meta: { title: '客户管理', icon: 'UserFilled', permissionCode: 'shop:customer:list' },
      },
      {
        path: 'order',
        name: 'ShopOrder',
        meta: { title: '订单管理', icon: 'Document' },
        redirect: '/shop/order/list',
        children: [
          {
            path: 'list',
            name: 'ShopOrderList',
            component: () => import('@/views/shop/order/index.vue'),
            meta: { title: '订单列表', icon: 'List', permissionCode: 'shop:order:list' },
          },
        ],
      },
      {
        path: 'finance',
        name: 'ShopFinance',
        meta: { title: '财务管理', icon: 'Money' },
        redirect: '/shop/finance/account',
        children: [
          {
            path: 'account',
            name: 'ShopFinanceAccount',
            component: () => import('@/views/shop/finance/account/index.vue'),
            meta: { title: '收支账户', icon: 'Wallet', permissionCode: 'shop:finance:account:list' },
          },
          {
            path: 'category',
            name: 'ShopFinanceCategory',
            component: () => import('@/views/shop/finance/category/index.vue'),
            meta: { title: '收支分类', icon: 'List', permissionCode: 'shop:finance:category:list' },
          },
          {
            path: 'record',
            name: 'ShopFinanceRecord',
            component: () => import('@/views/shop/finance/record/index.vue'),
            meta: { title: '收支记录', icon: 'Document', permissionCode: 'shop:finance:record:list' },
          },
          {
            path: 'report',
            name: 'ShopFinanceReport',
            component: () => import('@/views/shop/finance/report/index.vue'),
            meta: { title: '财务报表', icon: 'DataAnalysis', permissionCode: 'shop:finance:report:list' },
          },
        ],
      },
    ],
  },
]

export default shopRoutes
