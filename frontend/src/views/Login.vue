<template>
  <div class="auth-page">
    <div class="auth-card">
      <!-- 左侧品牌区 -->
      <div class="auth-left">
        <div class="brand">
          <svg viewBox="0 0 24 24" class="brand-icon"><path d="M14.06 9.02l.92.92L5.92 19H5v-.92l9.06-9.06M17.66 3c-.25 0-.51.1-.7.29l-1.83 1.83 3.75 3.75 1.83-1.83a.996.996 0 000-1.41l-2.34-2.34c-.2-.2-.45-.29-.71-.29zm-3.6 3.19L3 17.25V21h3.75L17.81 9.94l-3.75-3.75z" fill="currentColor"/></svg>
          <h1>OnlineJudge</h1>
        </div>
        <p class="brand-desc">在线编程判题平台</p>
        <div class="brand-features">
          <div class="feature-item">
            <span class="feature-icon">⚡</span>
            <span>实时代码评测</span>
          </div>
          <div class="feature-item">
            <span class="feature-icon">📊</span>
            <span>排行榜与提交记录</span>
          </div>
          <div class="feature-item">
            <span class="feature-icon">🎯</span>
            <span>多分类题库</span>
          </div>
        </div>
      </div>

      <!-- 右侧表单区 -->
      <div class="auth-right">
        <div class="form-wrapper">
          <h2>欢迎回来</h2>
          <p class="form-subtitle">登录你的账号</p>

          <form @submit.prevent="doLogin" class="auth-form">
            <div class="field">
              <label>用户名</label>
              <input v-model="form.username" type="text" placeholder="请输入用户名" class="input" required />
            </div>
            <div class="field">
              <label>密码</label>
              <input v-model="form.password" type="password" placeholder="请输入密码" class="input" required />
            </div>
            <div class="field-row">
              <label class="checkbox-label">
                <input type="checkbox" v-model="remember" /> 记住密码
              </label>
            </div>
            <div v-if="msg" class="form-msg" :class="{ error: isError }">{{ msg }}</div>
            <button type="submit" class="btn btn-primary btn-block" :disabled="loading">
              {{ loading ? '登录中...' : '登录' }}
            </button>
          </form>

          <p class="auth-link">
            没有账号？<router-link to="/register">去注册</router-link>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'

const router = useRouter()
const loading = ref(false)
const msg = ref('')
const isError = ref(false)
const remember = ref(false)
const form = reactive({ username: '', password: '' })

onMounted(() => {
  const savedUser = localStorage.getItem('oj_remember_user')
  const savedPass = localStorage.getItem('oj_remember_pass')
  if (savedUser) {
    form.username = savedUser
    form.password = savedPass || ''
    remember.value = true
  }
})

function doLogin() {
  loading.value = true; msg.value = ''; isError.value = false
  const fd = new URLSearchParams()
  fd.append('username', form.username)
  fd.append('password', form.password)
  api.post('/login', fd).then(res => {
    const d = res.data
    if (d.code === 200) {
      localStorage.setItem('oj_token', d.data.token)
      localStorage.setItem('oj_name', form.username)
      localStorage.setItem('oj_admin', d.data.isAdmin)
      if (remember.value) {
        localStorage.setItem('oj_remember_user', form.username)
        localStorage.setItem('oj_remember_pass', form.password)
      } else {
        localStorage.removeItem('oj_remember_user')
        localStorage.removeItem('oj_remember_pass')
      }
      router.push('/')
    } else {
      msg.value = d.msg || '登录失败'; isError.value = true
    }
  }).catch(() => { msg.value = '请求失败，请检查网络'; isError.value = true })
    .finally(() => { loading.value = false })
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 20px;
}
.auth-card {
  display: flex; width: 900px; max-width: 100%; min-height: 520px;
  background: #fff; border-radius: 16px; overflow: hidden;
  box-shadow: 0 20px 60px rgba(0,0,0,.2);
}

/* 左侧品牌 */
.auth-left {
  width: 380px; padding: 48px 40px; display: flex; flex-direction: column; justify-content: center;
  background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%); color: #fff;
}
.brand { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.brand-icon { width: 36px; height: 36px; }
.brand h1 { font-size: 24px; font-weight: 700; }
.brand-desc { font-size: 15px; opacity: .85; margin-bottom: 36px; }
.brand-features { display: flex; flex-direction: column; gap: 16px; }
.feature-item { display: flex; align-items: center; gap: 10px; font-size: 14px; opacity: .9; }
.feature-icon { font-size: 18px; }

/* 右侧表单 */
.auth-right { flex: 1; padding: 48px 40px; display: flex; align-items: center; }
.form-wrapper { width: 100%; max-width: 360px; margin: 0 auto; }
.form-wrapper h2 { font-size: 24px; font-weight: 700; margin-bottom: 4px; }
.form-subtitle { color: #6b7280; font-size: 14px; margin-bottom: 32px; }

.auth-form { display: flex; flex-direction: column; gap: 20px; }
.field label { display: block; font-size: 13px; font-weight: 600; color: #374151; margin-bottom: 6px; }
.input {
  width: 100%; padding: 10px 14px; border: 1px solid #e5e7eb; border-radius: 8px;
  font-size: 14px; transition: border .2s; background: #fff; color: #111827;
  font-family: inherit;
}
.input:focus { outline: none; border-color: #4f46e5; box-shadow: 0 0 0 3px rgba(79,70,229,.1); }

.field-row { display: flex; align-items: center; justify-content: space-between; }
.checkbox-label { display: flex; align-items: center; gap: 6px; font-size: 13px; color: #6b7280; cursor: pointer; }

.btn {
  padding: 10px 20px; border: none; border-radius: 8px; font-size: 14px; font-weight: 600;
  cursor: pointer; transition: all .2s; font-family: inherit;
}
.btn:disabled { opacity: .5; cursor: not-allowed; }
.btn-primary { background: #4f46e5; color: #fff; }
.btn-primary:hover:not(:disabled) { background: #4338ca; }
.btn-block { width: 100%; }

.form-msg { font-size: 13px; color: #059669; }
.form-msg.error { color: #dc2626; }

.auth-link { text-align: center; margin-top: 24px; font-size: 14px; color: #6b7280; }
.auth-link a { color: #4f46e5; text-decoration: none; font-weight: 500; }
.auth-link a:hover { text-decoration: underline; }

@media (max-width: 768px) {
  .auth-left { display: none; }
  .auth-card { min-height: auto; }
  .auth-right { padding: 32px 24px; }
}
</style>
