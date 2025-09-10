import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

// 用户信息接口
export interface UserInfo {
  id: number
  username: string
  nickname?: string
  email?: string
  avatar?: string
  roles?: string[]
  permissions?: string[]
}

// 登录请求接口
export interface LoginRequest {
  username: string
  password: string
  rememberMe?: boolean
}

// 登录响应接口
export interface LoginResponse {
  token: string
  user: UserInfo
  expiresIn?: number
}

// API 响应接口
export interface ApiResponse<T = any> {
  code: number
  data: T
  msg: string
}

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref<string>('')
  const userInfo = ref<UserInfo | null>(null)
  const isLoggedIn = computed(() => !!token.value && !!userInfo.value)

  // 从本地存储恢复状态
  const restoreFromStorage = () => {
    const savedToken = localStorage.getItem('sugar_token')
    const savedUserInfo = localStorage.getItem('sugar_user_info')
    
    if (savedToken) {
      token.value = savedToken
    }
    
    if (savedUserInfo) {
      try {
        userInfo.value = JSON.parse(savedUserInfo)
      } catch (error) {
        console.error('解析用户信息失败:', error)
        localStorage.removeItem('sugar_user_info')
      }
    }
  }

  // 保存到本地存储
  const saveToStorage = () => {
    if (token.value) {
      localStorage.setItem('sugar_token', token.value)
    }
    if (userInfo.value) {
      localStorage.setItem('sugar_user_info', JSON.stringify(userInfo.value))
    }
  }

  // 清除本地存储
  const clearStorage = () => {
    localStorage.removeItem('sugar_token')
    localStorage.removeItem('sugar_user_info')
    sessionStorage.removeItem('sugar_token')
    sessionStorage.removeItem('sugar_user_info')
  }

  // 设置 token
  const setToken = (newToken: string) => {
    token.value = newToken
    saveToStorage()
  }

  // 设置用户信息
  const setUserInfo = (info: UserInfo) => {
    userInfo.value = info
    saveToStorage()
  }

  // 登录
  const login = async (loginData: LoginRequest): Promise<void> => {
    try {
      // 这里使用模拟登录，实际项目中应该调用真实的登录接口
      const response = await mockLogin(loginData)
      
      if (response.code === 0) {
        setToken(response.data.token)
        setUserInfo(response.data.user)
        
        // 如果选择记住我，则保存到 localStorage，否则保存到 sessionStorage
        if (loginData.rememberMe) {
          saveToStorage()
        } else {
          sessionStorage.setItem('sugar_token', response.data.token)
          sessionStorage.setItem('sugar_user_info', JSON.stringify(response.data.user))
        }
      } else {
        throw new Error(response.msg || '登录失败')
      }
    } catch (error) {
      console.error('登录失败:', error)
      throw error
    }
  }

  // 登出
  const logout = async (): Promise<void> => {
    try {
      // 这里可以调用登出接口
      // await api.logout()
      
      // 清除状态
      token.value = ''
      userInfo.value = null
      clearStorage()
    } catch (error) {
      console.error('登出失败:', error)
      // 即使登出接口失败，也要清除本地状态
      token.value = ''
      userInfo.value = null
      clearStorage()
    }
  }

  // 刷新用户信息
  const refreshUserInfo = async (): Promise<void> => {
    if (!token.value) return
    
    try {
      // 这里应该调用获取用户信息的接口
      // const response = await api.getUserInfo()
      // setUserInfo(response.data)
    } catch (error) {
      console.error('刷新用户信息失败:', error)
      // 如果刷新失败，可能是 token 过期，需要重新登录
      await logout()
      throw error
    }
  }

  // 检查权限
  const hasPermission = (permission: string): boolean => {
    if (!userInfo.value?.permissions) return false
    return userInfo.value.permissions.includes(permission)
  }

  // 检查角色
  const hasRole = (role: string): boolean => {
    if (!userInfo.value?.roles) return false
    return userInfo.value.roles.includes(role)
  }

  // 初始化时恢复状态
  restoreFromStorage()

  return {
    // 状态
    token,
    userInfo,
    isLoggedIn,
    
    // 方法
    setToken,
    setUserInfo,
    login,
    logout,
    refreshUserInfo,
    hasPermission,
    hasRole,
    clearStorage
  }
})

// 模拟登录接口（实际项目中应该替换为真实的 API 调用）
const mockLogin = async (loginData: LoginRequest): Promise<ApiResponse<LoginResponse>> => {
  // 模拟网络延迟
  await new Promise(resolve => setTimeout(resolve, 1000))
  
  // 模拟登录验证
  if (loginData.username === 'admin' && loginData.password === '123456') {
    return {
      code: 0,
      msg: '登录成功',
      data: {
        token: 'mock_token_' + Date.now(),
        user: {
          id: 1,
          username: 'admin',
          nickname: '管理员',
          email: 'admin@sugar.com',
          avatar: '',
          roles: ['admin'],
          permissions: ['*']
        },
        expiresIn: 7200
      }
    }
  } else if (loginData.username === 'user' && loginData.password === '123456') {
    return {
      code: 0,
      msg: '登录成功',
      data: {
        token: 'mock_token_' + Date.now(),
        user: {
          id: 2,
          username: 'user',
          nickname: '普通用户',
          email: 'user@sugar.com',
          avatar: '',
          roles: ['user'],
          permissions: ['read', 'write']
        },
        expiresIn: 7200
      }
    }
  } else {
    return {
      code: 1,
      msg: '用户名或密码错误',
      data: null as any
    }
  }
}