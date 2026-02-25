import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import InstancesView from '../views/InstancesView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'login',
      component: LoginView,
    },
    {
      path: '/instances',
      name: 'instances',
      component: InstancesView,
    },
    {
      path: '/instances/new',
      name: 'create-instance',
      component: () => import('../views/CreateInstanceView.vue'),
    },
    {
      path: '/instances/:id',
      name: 'instance-detail',
      component: () => import('../views/InstanceDetailView.vue'),
    },
    {
      path: '/audit-logs',
      name: 'audit-logs',
      component: () => import('../views/AuditLogsView.vue'),
    },
    {
      path: '/service-logs',
      name: 'service-logs',
      component: () => import('../views/ServiceLogsView.vue'),
    },
  ],
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('jwt_token')

  if (!token && to.name !== 'login') {
    next({ name: 'login' })
    return
  }

  // Optional: if already logged in, prevent going back to login page
  if (token && to.name === 'login') {
    next({ name: 'instances' })
    return
  }

  next()
})

export default router
