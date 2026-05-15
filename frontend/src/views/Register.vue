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
          <h2>创建账号</h2>
          <p class="form-subtitle">注册后开始刷题之旅</p>

          <form @submit.prevent="doRegister" class="auth-form">
            <div class="field">
              <label>用户名</label>
              <input v-model="form.name" type="text" placeholder="请输入用户名" class="input" required />
            </div>
            <div class="field">
              <label>密码</label>
              <input v-model="form.password" type="password" placeholder="请输入密码" class="input" required />
            </div>
            <div class="field">
              <label>邮箱</label>
              <input v-model="form.mail" type="email" placeholder="请输入邮箱" class="input" required />
            </div>
            <div class="field">
              <label>手机号</label>
              <input v-model="form.phone" type="tel" placeholder="请输入手机号" maxlength="11" class="input" required />
            </div>
            <div class="field">
              <label>验证码</label>
              <div class="code-row">
                <input v-model="form.code" type="text" placeholder="请输入验证码" maxlength="6" class="input" required />
                <button type="button" class="btn btn-outline" @click="doSendCode" :disabled="countdown > 0">
                  {{ countdown > 0 ? countdown + 's' : '发送验证码' }}
                </button>
              </div>
            </div>
            <div v-if="msg" class="form-msg" :class="{ error: isError }">{{ msg }}</div>
            <button type="submit" class="btn btn-primary btn-block" :disabled="loading">
              {{ loading ? '注册中...' : '注册' }}
            </button>
          </form>

          <p class="auth-link">
            已有账号？<router-link to="/login">去登录</router-link>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'

const router = useRouter()
const loading = ref(false)
const msg = ref('')
const isError = ref(false)
const countdown = ref(0)
const form = reactive({ name: '', password: '', mail: '', phone: '', code: '' })

function doSendCode() {
  if (!form.mail) { msg.value = '请先输入邮箱'; isError.value = true; return }
  msg.value = ''; isError.value = false
  const fd = new URLSearchParams()
  fd.append('mail', form.mail)
  api.post('/sendcode', fd).then(res => {
    const d = res.data
    if (d.code === 200) {
      msg.value = '验证码已发送'; isError.value = false
      countdown.value = 60
      const t = setInterval(() => { countdown.value--; if (countdown.value <= 0) clearInterval(t) }, 1000)
    } else {
      msg.value = d.msg || '发送失败'; isError.value = true
    }
  }).catch(() => { msg.value = '请求失败'; isError.value = true })
}

function doRegister() {
  loading.value = true; msg.value = ''; isError.value = false
  const fd = new URLSearchParams()
  fd.append('name', form.name)
  fd.append('password', form.password)
  fd.append('mail', form.mail)
  fd.append('phone', form.phone)
  fd.append('code', form.code)
  api.post('/register', fd).then(res => {
    const d = res.data
    if (d.code === 200) {
      localStorage.setItem('oj_token', d.data.token)
      localStorage.setItem('oj_name', form.name)
      localStorage.setItem('oj_admin', '0')
      router.push('/')
    } else {
      msg.value = d.msg || '注册失败'; isError.value = true
    }
  }).catch(() => { msg.value = '请求失败'; isError.value = true })
    .finally(() => { loading.value = false })
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 20px;
}
.auth-card {
  display: flex; width: 900px; max-width: 100%; min-height: 580px;
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
.auth-right { flex: 1; padding: 40px; display: flex; align-items: center; overflow-y: auto; }
.form-wrapper { width: 100%; max-width: 360px; margin: 0 auto; }
.form-wrapper h2 { font-size: 24px; font-weight: 700; margin-bottom: 4px; }
.form-subtitle { color: #6b7280; font-size: 14px; margin-bottom: 28px; }

.auth-form { display: flex; flex-direction: column; gap: 16px; }
.field label { display: block; font-size: 13px; font-weight: 600; color: #374151; margin-bottom: 6px; }
.input {
  width: 100%; padding: 10px 14px; border: 1px solid #e5e7eb; border-radius: 8px;
  font-size: 14px; transition: border .2s; background: #fff; color: #111827;
  font-family: inherit;
}
.input:focus { outline: none; border-color: #4f46e5; box-shadow: 0 0 0 3px rgba(79,70,229,.1); }

.code-row { display: flex; gap: 12px; }
.code-row .input { flex: 1; }

.btn {
  padding: 10px 20px; border: none; border-radius: 8px; font-size: 14px; font-weight: 600;
  cursor: pointer; transition: all .2s; font-family: inherit;
}
.btn:disabled { opacity: .5; cursor: not-allowed; }
.btn-primary { background: #4f46e5; color: #fff; }
.btn-primary:hover:not(:disabled) { background: #4338ca; }
.btn-outline { background: transparent; color: #374151; border: 1px solid #e5e7eb; white-space: nowrap; }
.btn-outline:hover:not(:disabled) { background: #f3f4f6; border-color: #9ca3af; }
.btn-block { width: 100%; }

.form-msg { font-size: 13px; color: #059669; }
.form-msg.error { color: #dc2626; }

.auth-link { text-align: center; margin-top: 20px; font-size: 14px; color: #6b7280; }
.auth-link a { color: #4f46e5; text-decoration: none; font-weight: 500; }
.auth-link a:hover { text-decoration: underline; }

@media (max-width: 768px) {
  .auth-left { display: none; }
  .auth-card { min-height: auto; }
  .auth-right { padding: 32px 24px; }
}
</style>
