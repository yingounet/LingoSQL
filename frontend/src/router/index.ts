import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // 公开页面（无需认证）
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('@/views/Register.vue'),
      meta: { requiresAuth: false },
    },
    // 需要认证的页面（使用 MainLayout）
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue'),
          meta: { 
            title: 'Dashboard',
            breadcrumb: '首页'
          },
        },
        {
          path: 'query',
          name: 'Query',
          component: () => import('@/views/Query.vue'),
          meta: { 
            title: 'SQL Editor',
            breadcrumb: 'SQL 查询',
            keepAlive: true
          },
        },
        {
          path: 'schema',
          name: 'Schema',
          component: () => import('@/views/Schema.vue'),
          meta: { 
            title: 'Table Schema',
            breadcrumb: '表结构',
            keepAlive: true
          },
        },
        {
          path: 'rowdata',
          name: 'RowData',
          component: () => import('@/views/RowData.vue'),
          meta: { 
            title: 'Row Data',
            breadcrumb: '表数据',
            keepAlive: true
          },
        },
        {
          path: 'connections',
          redirect: '/',
        },
        {
          path: 'database',
          name: 'Database',
          component: () => import('@/views/Database.vue'),
          meta: { 
            title: 'Databases',
            breadcrumb: '数据库',
            activeNav: '/database'
          },
        },
        {
          path: 'history',
          name: 'History',
          component: () => import('@/views/History.vue'),
          meta: { 
            title: 'History',
            breadcrumb: '查询历史'
          },
        },
        {
          path: 'favorites',
          name: 'Favorites',
          component: () => import('@/views/Favorites.vue'),
          meta: { 
            title: 'Favorites',
            breadcrumb: '收藏'
          },
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/Settings.vue'),
          meta: { 
            title: '个人设置',
            breadcrumb: '个人设置'
          },
        },
      ],
    },
  ],
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()
  const requiresAuth = to.meta.requiresAuth !== false

  if (requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (!requiresAuth && authStore.isAuthenticated && (to.path === '/login' || to.path === '/register')) {
    next('/')
  } else {
    next()
  }
})

// 路由后置守卫 - 更新页面标题
router.afterEach((to) => {
  const title = to.meta.title as string
  document.title = title ? `${title} - LingoSQL` : 'LingoSQL'
})

export default router
