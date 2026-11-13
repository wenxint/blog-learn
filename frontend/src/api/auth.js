import request from './request'

export function login(username, password) {
  return request.post('/auth/login', { username, password })
}

export function fetchProfile() {
  return request.get('/auth/profile')
}

export function changePassword(oldPassword, newPassword) {
  return request.put('/auth/password', {
    old_password: oldPassword,
    new_password: newPassword,
  })
}
