import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/dashboard'
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: {
        title: '登录',
        requiresAuth: false
      }
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('@/views/Dashboard.vue'),
      meta: {
        title: '仪表板',
        requiresAuth: true
      }
    },
    {
      path: '/sugar',
      name: 'SugarApp',
      component: () => import('@/views/SugarApp.vue'),
      meta: {
        title: 'Sugar Analytics',
        requiresAuth: true,
        keepAlive: true
      }
    }
  ],
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  // 设置页面标题
  if (to.meta?.title) {
    document.title = to.meta.title as string
  }
  
  // 检查是否需要登录
  if (to.meta?.requiresAuth && !userStore.isLoggedIn) {
    next('/login')
  } else if (to.path === '/login' && userStore.isLoggedIn) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
