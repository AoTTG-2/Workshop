import { reactive } from 'vue'

interface AuthState {
  userId: string
  userRoles: string[]
}

export const auth = reactive<AuthState>({
  userId: '',
  userRoles: []
})

export function initAuth() {
  const storedUserId = localStorage.getItem('userId')
  const storedRoles = localStorage.getItem('userRoles')

  if (storedUserId) {
    auth.userId = storedUserId
  }
  if (storedRoles) {
    try {
      auth.userRoles = JSON.parse(storedRoles)
    } catch {
      auth.userRoles = []
    }
  }
}

export function setAuth(userId: string, userRoles: string[]) {
  auth.userId = userId
  auth.userRoles = userRoles
  localStorage.setItem('userId', userId)
  localStorage.setItem('userRoles', JSON.stringify(userRoles))
}

export function logout() {
  auth.userId = ''
  auth.userRoles = []
  localStorage.removeItem('userId')
  localStorage.removeItem('userRoles')
}
