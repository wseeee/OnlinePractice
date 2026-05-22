<template>
  <div class="oj-app">
    <!-- ==================== 顶部导航栏 ==================== -->
    <header class="oj-header">
      <div class="header-inner">
        <div class="logo" @click="currentView = 'problems'">
          <svg viewBox="0 0 24 24" class="logo-icon"><path d="M14.06 9.02l.92.92L5.92 19H5v-.92l9.06-9.06M17.66 3c-.25 0-.51.1-.7.29l-1.83 1.83 3.75 3.75 1.83-1.83a.996.996 0 000-1.41l-2.34-2.34c-.2-.2-.45-.29-.71-.29zm-3.6 3.19L3 17.25V21h3.75L17.81 9.94l-3.75-3.75z" fill="currentColor"/></svg>
          <span>OnlineJudge</span>
        </div>
        <nav class="header-nav">
          <a :class="{ active: currentView === 'problems' }" @click="switchView('problems')">题库</a>
          <a :class="{ active: currentView === 'rank' }" @click="switchView('rank')">排行榜</a>
          <a :class="{ active: currentView === 'contest' }" @click="switchView('contest')">比赛</a>
          <a :class="{ active: currentView === 'submit' }" @click="switchView('submit')">提交记录</a>
          <a v-if="isAdmin" :class="{ active: currentView === 'admin' }" @click="switchView('admin')">管理</a>
        </nav>
        <div class="header-right">
          <template v-if="token">
            <span class="user-name">{{ userName }}</span>
            <button class="btn btn-outline" @click="logout">退出</button>
          </template>
          <template v-else>
            <button class="btn btn-primary" @click="router.push('/login')">登录</button>
            <button class="btn btn-outline" @click="router.push('/register')">注册</button>
          </template>
        </div>
      </div>
    </header>

    <!-- ==================== 主内容区 ==================== -->
    <main class="oj-main">
      <div class="container">

        <!-- ===== 题库页 ===== -->
        <section v-if="currentView === 'problems'" class="view-problems">
          <div class="page-header">
            <h2>题库</h2>
            <p class="subtitle">挑战自己，提升编程能力</p>
          </div>
          <div class="search-bar">
            <input v-model="problemSearch.keyword" placeholder="搜索题目..." @keyup.enter="fetchProblems" class="input" />
            <select v-model="problemSearch.category" @change="fetchProblems" class="input select">
              <option value="">全部分类</option>
              <option v-for="cat in categories" :key="cat.identity" :value="cat.identity">{{ cat.name }}</option>
            </select>
            <button class="btn btn-primary" @click="fetchProblems">搜索</button>
          </div>
          <div v-if="problemLoading" class="state-box"><span class="spinner"></span> 加载中...</div>
          <div v-else-if="problemError" class="state-box error">{{ problemError }}</div>
          <div v-else-if="problems.length === 0" class="state-box empty">暂无题目数据</div>
          <div v-else class="problem-table-wrap">
            <table class="data-table">
              <thead><tr><th>状态</th><th>标题</th><th>通过率</th><th>提交数</th></tr></thead>
              <tbody>
                <tr v-for="p in problems" :key="p.identity" @click="openProblemDetail(p)" class="clickable">
                  <td><span class="status-dot" :class="{ passed: p.passed }"></span></td>
                  <td>{{ p.title }}</td>
                  <td>{{ p.submit_num ? Math.round(p.pass_num / p.submit_num * 100) : 0 }}%</td>
                  <td>{{ p.submit_num }}</td>
                </tr>
              </tbody>
            </table>
            <div class="pagination">
              <button :disabled="problemPage <= 1" @click="problemPage--; fetchProblems()">上一页</button>
              <span>第 {{ problemPage }} 页 / 共 {{ Math.ceil(problemTotal / problemSize) || 1 }} 页</span>
              <button :disabled="problemPage >= Math.ceil(problemTotal / problemSize)" @click="problemPage++; fetchProblems()">下一页</button>
            </div>
          </div>
        </section>

        <!-- ===== 比赛列表页 ===== -->
        <section v-if="currentView === 'contest'" class="view-problems">
          <div class="page-header">
            <h2>比赛</h2>
            <p class="subtitle">参加比赛，挑战自我</p>
          </div>
          <div v-if="contestLoading" class="state-box"><span class="spinner"></span> 加载中...</div>
          <div v-else-if="contestError" class="state-box error">{{ contestError }}</div>
          <div v-else-if="contests.length === 0" class="state-box empty">暂无比赛</div>
          <div v-else class="problem-table-wrap">
            <table class="data-table">
              <thead><tr><th>比赛</th><th>开始时间</th><th>结束时间</th><th>状态</th><th>参赛人数</th></tr></thead>
              <tbody>
                <tr v-for="c in contests" :key="c.identity" @click="openContestDetail(c)" class="clickable">
                  <td>{{ c.title }}</td>
                  <td>{{ formatTime(c.start_at) }}</td>
                  <td>{{ formatTime(c.end_at) }}</td>
                  <td><span class="status-badge" :class="contestStatusClass(c.contest_status)">{{ contestStatusText(c.contest_status) }}</span></td>
                  <td>{{ c.participant_count }}</td>
                </tr>
              </tbody>
            </table>
            <div class="pagination">
              <button :disabled="contestPage <= 1" @click="contestPage--; fetchContests()">上一页</button>
              <span>第 {{ contestPage }} 页 / 共 {{ Math.ceil(contestTotal / contestSize) || 1 }} 页</span>
              <button :disabled="contestPage >= Math.ceil(contestTotal / contestSize)" @click="contestPage++; fetchContests()">下一页</button>
            </div>
          </div>
        </section>

        <!-- ===== 比赛详情页 ===== -->
        <section v-if="currentView === 'contest-detail'" class="view-detail">
          <button class="btn btn-text" @click="currentView = 'contest'; fetchContests()">← 返回比赛列表</button>
          <div v-if="contestDetailLoading" class="state-box"><span class="spinner"></span> 加载中...</div>
          <div v-else-if="contestDetailError" class="state-box error">{{ contestDetailError }}</div>
          <div v-else class="detail-layout">
            <div class="detail-left">
              <h2>{{ contestDetail.contest.title }}</h2>
              <div class="detail-meta">
                <span>开始: {{ formatTime(contestDetail.contest.start_at) }}</span>
                <span>结束: {{ formatTime(contestDetail.contest.end_at) }}</span>
                <span>参赛人数: {{ contestDetail.participant_count }}</span>
                <span class="status-badge" :class="contestStatusClass(contestDetail.status)">{{ contestStatusText(contestDetail.status) }}</span>
              </div>
              <div v-if="contestDetail.contest.description" class="detail-content" style="margin-top:16px">
                <pre>{{ contestDetail.contest.description }}</pre>
              </div>
              <h3 style="margin-top:24px">题目列表</h3>
              <table class="data-table" style="margin-top:8px">
                <thead><tr><th>#</th><th>题目</th><th>满分</th><th>我的得分</th></tr></thead>
                <tbody>
                  <tr v-for="(cp, idx) in contestDetail.contest.contest_problems" :key="cp.id"
                      @click="openContestProblem(cp)" :class="{ clickable: true, 'selected-row': selectedContestProblem?.problem_identity === cp.problem_identity }">
                    <td>{{ idx + 1 }}</td>
                    <td>{{ cp.problem_basic ? cp.problem_basic.title : cp.problem_identity }}</td>
                    <td>{{ cp.score }}</td>
                    <td>{{ contestMyScores[cp.problem_identity] ?? '-' }}</td>
                  </tr>
                </tbody>
              </table>
              <div v-if="contestDetail.contest.contest_type === 2 && contestDetail.status === 0" style="margin-top:16px">
                <button class="btn btn-primary" @click="joinContest" :disabled="contestJoining">
                  {{ contestJoining ? '报名中...' : '报名参加' }}
                </button>
                <span v-if="contestJoinMsg" style="margin-left:8px; font-size:14px">{{ contestJoinMsg }}</span>
              </div>
            </div>
            <div class="detail-right">
              <div class="right-tabs">
                <a :class="{ active: contestTab === 'submit' }" @click="contestTab = 'submit'">我的提交</a>
                <a :class="{ active: contestTab === 'rank' }" @click="contestTab = 'rank'; fetchContestRank()">排行榜</a>
              </div>
              <div v-if="contestTab === 'submit'">
                <div v-if="selectedContestProblem" style="margin-bottom:8px; font-size:14px">
                  <strong>当前题目：</strong>{{ selectedContestProblem.problem_basic?.title }}
                </div>
                <textarea v-model="contestCode" class="code-editor" placeholder="在此编写 Go 代码..." rows="12"></textarea>
                <div style="margin-top:8px; display:flex; gap:8px">
                  <button class="btn btn-primary" @click="submitContestCode" :disabled="contestSubmitting || !selectedContestProblem">
                    {{ contestSubmitting ? '提交中...' : '提交代码' }}
                  </button>
                  <span v-if="contestSubmitMsg" :style="{ color: contestSubmitColor, fontSize: '14px', alignSelf: 'center' }">{{ contestSubmitMsg }}</span>
                </div>
                <div v-if="contestSubmits.length > 0" style="margin-top:16px">
                  <table class="data-table">
                    <thead><tr><th>题目</th><th>得分</th><th>状态</th><th>时间</th></tr></thead>
                    <tbody>
                      <tr v-for="s in contestSubmits" :key="s.id">
                        <td>{{ s.problem_basic?.title ?? s.problem_identity }}</td>
                        <td>{{ s.score }}</td>
                        <td><span class="status-tag" :class="statusClass(s.status)">{{ statusLabel(s.status) }}</span></td>
                        <td>{{ formatTime(s.created_at) }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
              <div v-if="contestTab === 'rank'">
                <div v-if="contestRankLoading" class="state-box"><span class="spinner"></span> 加载中...</div>
                <div v-else-if="contestRanks.length === 0" class="state-box empty">暂无排名数据</div>
                <table v-else class="data-table">
                  <thead><tr><th>排名</th><th>用户</th><th>总分</th></tr></thead>
                  <tbody>
                    <tr v-for="(r, idx) in contestRanks" :key="r.user_identity">
                      <td><span class="rank-badge" :class="'rank-' + (idx + 1)">{{ idx + 1 }}</span></td>
                      <td>{{ r.user_name }}</td>
                      <td>{{ r.total_score }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>

        <!-- ===== 题目详情页 ===== -->
        <section v-if="currentView === 'detail'" class="view-detail">
          <button class="btn btn-text" @click="currentView = 'problems'">← 返回题库</button>
          <div v-if="detailLoading" class="state-box"><span class="spinner"></span> 加载中...</div>
          <div v-else-if="detailError" class="state-box error">{{ detailError }}</div>
          <div v-else class="detail-layout">
            <div class="detail-left">
              <h2>{{ detail.title }}</h2>
              <div class="detail-meta">
                <span>时间限制: {{ detail.max_runtime }}ms</span>
                <span>内存限制: {{ detail.max_mem }}KB</span>
                <span>提交: {{ detail.submit_num }}</span>
                <span>通过: {{ detail.pass_num }}</span>
              </div>
              <div class="detail-content" v-html="renderedContent"></div>
            </div>
            <div class="detail-right">
              <div class="editor-panel">
                <div class="panel-header">提交代码 (Go)</div>
                <textarea v-model="submitCode" class="code-editor" placeholder="// 在此编写你的 Go 代码..." spellcheck="false"></textarea>
                <button class="btn btn-primary btn-block" @click="doSubmit" :disabled="submitting || !token">
                  {{ submitting ? '提交中...' : token ? '提交代码' : '请先登录' }}
                </button>
                <div v-if="submitResult" class="submit-result" :class="submitResultClass">{{ submitResult }}</div>
              </div>
            </div>
          </div>
        </section>

        <!-- ===== 排行榜 ===== -->
        <section v-if="currentView === 'rank'" class="view-rank">
          <div class="page-header"><h2>排行榜</h2><p class="subtitle">刷题达人</p></div>
          <div v-if="rankLoading" class="state-box"><span class="spinner"></span> 加载中...</div>
          <div v-else-if="rankError" class="state-box error">{{ rankError }}</div>
          <div v-else-if="rankList.length === 0" class="state-box empty">暂无排名数据</div>
          <div v-else class="rank-table-wrap">
            <table class="data-table">
              <thead><tr><th>排名</th><th>用户名</th><th>通过数</th><th>提交数</th><th>通过率</th></tr></thead>
              <tbody>
                <tr v-for="(u, i) in rankList" :key="u.identity">
                  <td><span class="rank-badge" :class="'rank-' + (i + 1)">{{ i + 1 }}</span></td>
                  <td>{{ u.name }}</td>
                  <td>{{ u.pass_num }}</td>
                  <td>{{ u.submit_num }}</td>
                  <td>{{ u.submit_num ? Math.round(u.pass_num / u.submit_num * 100) : 0 }}%</td>
                </tr>
              </tbody>
            </table>
            <div class="pagination">
              <button :disabled="rankPage <= 1" @click="rankPage--; fetchRank()">上一页</button>
              <span>第 {{ rankPage }} 页 / 共 {{ Math.ceil(rankTotal / rankSize) || 1 }} 页</span>
              <button :disabled="rankPage >= Math.ceil(rankTotal / rankSize)" @click="rankPage++; fetchRank()">下一页</button>
            </div>
          </div>
        </section>

        <!-- ===== 提交记录 ===== -->
        <section v-if="currentView === 'submit'" class="view-submit">
          <div class="page-header"><h2>提交记录</h2></div>
          <div class="search-bar">
            <input v-model="submitSearch.problem" placeholder="题目ID" class="input" />
            <input v-model="submitSearch.user" placeholder="用户ID" class="input" />
            <select v-model="submitSearch.status" class="input select">
              <option value="">全部状态</option>
              <option value="1">答案正确</option>
              <option value="2">答案错误</option>
              <option value="3">运行超时</option>
              <option value="4">运行超内存</option>
              <option value="5">编译错误</option>
              <option value="6">无效代码</option>
            </select>
            <button class="btn btn-primary" @click="fetchSubmitList">搜索</button>
          </div>
          <div v-if="submitLoading" class="state-box"><span class="spinner"></span> 加载中...</div>
          <div v-else-if="submitError" class="state-box error">{{ submitError }}</div>
          <div v-else-if="submitList.length === 0" class="state-box empty">暂无提交记录</div>
          <div v-else>
            <table class="data-table">
              <thead><tr><th>题目</th><th>用户</th><th>状态</th><th>时间</th></tr></thead>
              <tbody>
                <tr v-for="s in submitList" :key="s.identity">
                  <td>{{ s.problem_basic?.title || s.problem_identity }}</td>
                  <td>{{ s.user_basic?.name || s.user_identity }}</td>
                  <td><span class="status-tag" :class="statusClass(s.status)">{{ statusLabel(s.status) }}</span></td>
                  <td>{{ formatTime(s.CreatedAt) }}</td>
                </tr>
              </tbody>
            </table>
            <div class="pagination">
              <button :disabled="submitPage <= 1" @click="submitPage--; fetchSubmitList()">上一页</button>
              <span>第 {{ submitPage }} 页 / 共 {{ Math.ceil(submitTotal / submitSize) || 1 }} 页</span>
              <button :disabled="submitPage >= Math.ceil(submitTotal / submitSize)" @click="submitPage++; fetchSubmitList()">下一页</button>
            </div>
          </div>
        </section>

        <!-- ===== 管理员面板 ===== -->
        <section v-if="currentView === 'admin' && isAdmin" class="view-admin">
          <div class="page-header"><h2>管理面板</h2></div>
          <div class="admin-tabs">
            <button :class="{ active: adminTab === 'categories' }" @click="adminTab = 'categories'">分类管理</button>
            <button :class="{ active: adminTab === 'problem-create' }" @click="adminTab = 'problem-create'">新建题目</button>
            <button :class="{ active: adminTab === 'contest-admin' }" @click="adminTab = 'contest-admin'; fetchAdminContests()">比赛管理</button>
          </div>

          <!-- 分类管理 -->
          <div v-if="adminTab === 'categories'" class="admin-section">
            <div class="search-bar">
              <input v-model="catSearch.keyword" placeholder="搜索分类..." class="input" @keyup.enter="fetchCategories" />
              <button class="btn btn-primary" @click="fetchCategories">搜索</button>
              <button class="btn btn-success" @click="showCatForm = !showCatForm; editingCat = null; catForm = { name: '', parent_id: 0 }">
                {{ showCatForm ? '取消' : '新增分类' }}
              </button>
            </div>
            <div v-if="showCatForm" class="form-card">
              <h4>{{ editingCat ? '编辑分类' : '新建分类' }}</h4>
              <input v-model="catForm.name" placeholder="分类名称" class="input" />
              <input v-model.number="catForm.parent_id" placeholder="父级ID (可选)" type="number" class="input" />
              <button class="btn btn-primary" @click="saveCategory">{{ editingCat ? '更新' : '创建' }}</button>
            </div>
            <div v-if="catLoading" class="state-box"><span class="spinner"></span> 加载中...</div>
            <table v-else-if="catList.length" class="data-table">
              <thead><tr><th>名称</th><th>父级ID</th><th>操作</th></tr></thead>
              <tbody>
                <tr v-for="c in catList" :key="c.identity">
                  <td>{{ c.name }}</td>
                  <td>{{ c.parent_id }}</td>
                  <td>
                    <button class="btn btn-sm btn-outline" @click="editCategory(c)">编辑</button>
                    <button class="btn btn-sm btn-danger" @click="deleteCategory(c.identity)">删除</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- 新建/编辑题目 -->
          <div v-if="adminTab === 'problem-create'" class="admin-section">
            <div class="form-card">
              <h4>{{ editingProblem ? '编辑题目' : '新建题目' }}</h4>
              <input v-model="problemForm.title" placeholder="题目标题" class="input" />
              <textarea v-model="problemForm.content" placeholder="题目内容 (支持HTML)" class="input textarea"></textarea>
              <div class="form-row">
                <input v-model.number="problemForm.max_runtime" placeholder="最大运行时间(ms)" type="number" class="input" />
                <input v-model.number="problemForm.max_mem" placeholder="最大内存(KB)" type="number" class="input" />
              </div>
              <div class="form-group">
                <label>分类 (多选)</label>
                <div class="checkbox-group">
                  <label v-for="c in allCategories" :key="c.identity" class="checkbox-label">
                    <input type="checkbox" :value="c.identity" v-model="problemForm.category_ids" /> {{ c.name }}
                  </label>
                </div>
              </div>
              <div class="form-group">
                <label>测试用例</label>
                <div v-for="(tc, i) in problemForm.test_cases" :key="i" class="test-case-row">
                  <input v-model="tc.input" placeholder="输入" class="input" />
                  <input v-model="tc.output" placeholder="输出" class="input" />
                  <button class="btn btn-sm btn-danger" @click="problemForm.test_cases.splice(i, 1)">删除</button>
                </div>
                <button class="btn btn-sm btn-outline" @click="problemForm.test_cases.push({ input: '', output: '' })">+ 添加测试用例</button>
              </div>
              <div class="form-actions">
                <button class="btn btn-primary" @click="saveProblem" :disabled="problemSaving">
                  {{ problemSaving ? '保存中...' : editingProblem ? '更新题目' : '创建题目' }}
                </button>
                <button v-if="editingProblem" class="btn btn-text" @click="editingProblem = null; resetProblemForm()">取消编辑</button>
              </div>
              <div v-if="problemFormMsg" class="form-msg" :class="{ error: problemFormError }">{{ problemFormMsg }}</div>
            </div>
            <div class="search-bar" style="margin-top:24px">
              <input v-model="problemSearch2.keyword" placeholder="搜索题目..." class="input" @keyup.enter="fetchAdminProblems" />
              <button class="btn btn-primary" @click="fetchAdminProblems">搜索</button>
            </div>
            <table v-if="adminProblems.length" class="data-table">
              <thead><tr><th>标题</th><th>操作</th></tr></thead>
              <tbody>
                <tr v-for="p in adminProblems" :key="p.identity">
                  <td>{{ p.title }}</td>
                  <td>
                    <button class="btn btn-sm btn-outline" @click="loadProblemForEdit(p.identity)">编辑</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- 比赛管理 -->
          <div v-if="adminTab === 'contest-admin'" class="admin-section">
            <div class="search-bar">
              <button class="btn btn-success" @click="showContestForm = !showContestForm; editingContest = null; resetContestForm()">
                {{ showContestForm ? '取消' : '新建比赛' }}
              </button>
            </div>
            <div v-if="showContestForm" class="form-card">
              <h4>{{ editingContest ? '编辑比赛' : '新建比赛' }}</h4>
              <input v-model="contestForm.title" placeholder="比赛标题" class="input" />
              <textarea v-model="contestForm.description" placeholder="比赛描述" class="input textarea"></textarea>
              <div class="form-row">
                <div>
                  <label style="font-size:13px; color:var(--text-secondary)">开始时间</label>
                  <input v-model="contestForm.start_at" type="datetime-local" class="input" />
                </div>
                <div>
                  <label style="font-size:13px; color:var(--text-secondary)">结束时间</label>
                  <input v-model="contestForm.end_at" type="datetime-local" class="input" />
                </div>
              </div>
              <div class="form-row">
                <div>
                  <label style="font-size:13px; color:var(--text-secondary)">比赛类型</label>
                  <select v-model="contestForm.contest_type" class="input select">
                    <option :value="1">公开比赛</option>
                    <option :value="2">报名制</option>
                  </select>
                </div>
                <div v-if="contestForm.contest_type === 2">
                  <label style="font-size:13px; color:var(--text-secondary)">人数上限 (0=不限)</label>
                  <input v-model.number="contestForm.max_participants" type="number" class="input" />
                </div>
              </div>
              <div class="form-group">
                <label>选择题目</label>
                <div class="checkbox-group">
                  <label v-for="p in allProblemsForContest" :key="p.identity" class="checkbox-label">
                    <input type="checkbox" :value="p.identity" v-model="contestForm.problem_identities" /> {{ p.title }}
                  </label>
                </div>
              </div>
              <div class="form-actions">
                <button class="btn btn-primary" @click="saveContest" :disabled="contestSaving">
                  {{ contestSaving ? '保存中...' : editingContest ? '更新比赛' : '创建比赛' }}
                </button>
              </div>
              <div v-if="contestFormMsg" class="form-msg" :class="{ error: contestFormError }">{{ contestFormMsg }}</div>
            </div>
            <div v-if="adminContestLoading" class="state-box"><span class="spinner"></span> 加载中...</div>
            <table v-else-if="adminContests.length" class="data-table">
              <thead><tr><th>标题</th><th>类型</th><th>开始时间</th><th>操作</th></tr></thead>
              <tbody>
                <tr v-for="c in adminContests" :key="c.identity">
                  <td>{{ c.title }}</td>
                  <td>{{ c.contest_type === 1 ? '公开' : '报名制' }}</td>
                  <td>{{ formatTime(c.start_at) }}</td>
                  <td>
                    <button class="btn btn-sm btn-outline" @click="loadContestForEdit(c)">编辑</button>
                    <button class="btn btn-sm btn-danger" @click="deleteContest(c.identity)">删除</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

      </div>
    </main>

    <!-- ==================== Toast 通知 ==================== -->
    <transition name="toast-fade">
      <div v-if="toast.show" class="toast" :class="toast.type">{{ toast.msg }}</div>
    </transition>
  </div>
</template>

<script setup>
import { reactive, ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from './api'

const router = useRouter()

// ======================== 全局状态 ========================
const currentView = ref('problems')
const token = ref(localStorage.getItem('oj_token') || '')
const userName = ref(localStorage.getItem('oj_name') || '')
const isAdmin = ref(Number(localStorage.getItem('oj_admin')) === 1)

// Toast
const toast = reactive({ show: false, msg: '', type: 'info' })
let toastTimer = null
function showToast(msg, type = 'info') {
  toast.msg = msg; toast.type = type; toast.show = true
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toast.show = false }, 3000)
}

// ======================== 视图切换 ========================
function switchView(v) {
  currentView.value = v
  if (v === 'problems') fetchProblems()
  else if (v === 'rank') fetchRank()
  else if (v === 'submit') fetchSubmitList()
  else if (v === 'contest') fetchContests()
  else if (v === 'admin') { fetchCategories(); fetchAllCategories() }
}

// ======================== 登出 ========================
function logout() {
  token.value = ''; userName.value = ''; isAdmin.value = false
  localStorage.removeItem('oj_token'); localStorage.removeItem('oj_name'); localStorage.removeItem('oj_admin')
  router.push('/login')
}

// ======================== 题库 ========================
const problems = ref([])
const problemLoading = ref(false)
const problemError = ref('')
const problemTotal = ref(0)
const problemPage = ref(1)
const problemSize = 20
const problemSearch = reactive({ keyword: '', category: '' })

function fetchProblems() {
  problemLoading.value = true; problemError.value = ''
  api.get('/problem-list', {
    params: { page: problemPage.value, size: problemSize, keyword: problemSearch.keyword, category_identity: problemSearch.category }
  }).then(res => {
    const d = res.data
    if (d.code === 200) { problems.value = d.data.list; problemTotal.value = d.data.count }
    else problemError.value = d.msg || '获取失败'
  }).catch(() => { problemError.value = '请求失败' }).finally(() => { problemLoading.value = false })
}

// ======================== 题目详情 & 提交 ========================
const currentProblemId = ref('')
const detail = ref({})
const detailLoading = ref(false)
const detailError = ref('')
const submitCode = ref('')
const submitting = ref(false)
const submitResult = ref('')
const submitResultClass = ref('')

const renderedContent = computed(() => detail.value.content || '')

function openProblemDetail(p) {
  currentProblemId.value = p.identity
  detailLoading.value = true; detailError.value = ''; submitResult.value = ''
  api.get('/problem-detail', { params: { identity: p.identity } }).then(res => {
    const d = res.data
    if (d.code === 200) detail.value = d.data
    else detailError.value = d.msg || '获取失败'
  }).catch(() => { detailError.value = '请求失败' }).finally(() => { detailLoading.value = false })
  currentView.value = 'detail'
}

function doSubmit() {
  if (!submitCode.value.trim()) { showToast('请输入代码', 'error'); return }
  submitting.value = true; submitResult.value = ''
  api.post('/user/code-submit', submitCode.value, {
    params: { problem_identity: currentProblemId.value },
    headers: { 'Content-Type': 'text/plain' }
  }).then(res => {
    const d = res.data
    if (d.code === 200) {
      const status = d.data.status
      const labels = { 1: '答案正确', 2: '答案错误', 3: '运行超时', 4: '运行超内存', 5: '编译错误', 6: '无效代码' }
      const classes = { 1: 'success', 2: 'error', 3: 'error', 4: 'error', 5: 'error', 6: 'error' }
      submitResult.value = labels[status] || d.data.msg
      submitResultClass.value = classes[status] || 'error'
    } else {
      submitResult.value = d.msg || '提交失败'; submitResultClass.value = 'error'
    }
  }).catch(() => { submitResult.value = '提交请求失败'; submitResultClass.value = 'error' }).finally(() => { submitting.value = false })
}

// ======================== 排行榜 ========================
const rankList = ref([])
const rankLoading = ref(false)
const rankError = ref('')
const rankTotal = ref(0)
const rankPage = ref(1)
const rankSize = 20

function fetchRank() {
  rankLoading.value = true; rankError.value = ''
  api.get('/rank-list', { params: { page: rankPage.value, size: rankSize } }).then(res => {
    const d = res.data
    if (d.code === 200) { rankList.value = d.data.list; rankTotal.value = d.data.count }
    else rankError.value = d.msg || '获取失败'
  }).catch(() => { rankError.value = '请求失败' }).finally(() => { rankLoading.value = false })
}

// ======================== 提交记录 ========================
const submitList = ref([])
const submitLoading = ref(false)
const submitError = ref('')
const submitTotal = ref(0)
const submitPage = ref(1)
const submitSize = 20
const submitSearch = reactive({ problem: '', user: '', status: '' })

function fetchSubmitList() {
  submitLoading.value = true; submitError.value = ''
  api.get('/submit-list', {
    params: {
      page: submitPage.value, size: submitSize,
      problem_identity: submitSearch.problem, user_identity: submitSearch.user,
      status: submitSearch.status || undefined
    }
  }).then(res => {
    const d = res.data
    if (d.code === 200) { submitList.value = d.data.list; submitTotal.value = d.data.count }
    else submitError.value = d.msg || '获取失败'
  }).catch(() => { submitError.value = '请求失败' }).finally(() => { submitLoading.value = false })
}

function statusClass(s) {
  return { 1: 'success', 2: 'error', 3: 'warn', 4: 'warn', 5: 'error', 6: 'error' }[s] || ''
}
function statusLabel(s) {
  return { 1: 'AC', 2: 'WA', 3: 'TLE', 4: 'MLE', 5: 'CE', 6: '无效' }[s] || '未知'
}
function formatTime(t) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

// ======================== 分类管理 ========================
const categories = ref([])         // 公共分类列表
const allCategories = ref([])      // 管理用全量分类
const catLoading = ref(false)
const catList = ref([])
const catTotal = ref(0)
const catPage = ref(1)
const catSearch = reactive({ keyword: '' })
const showCatForm = ref(false)
const editingCat = ref(null)
const catForm = reactive({ name: '', parent_id: 0 })

function fetchCategories() {
  catLoading.value = true
  api.get('/admin/category-list', { params: { page: catPage.value, size: 20, keyword: catSearch.keyword } }).then(res => {
    const d = res.data
    if (d.code === 200) { catList.value = d.data.list; catTotal.value = d.data.count; categories.value = d.data.list }
  }).finally(() => { catLoading.value = false })
}

function fetchAllCategories() {
  api.get('/admin/category-list', { params: { page: 1, size: 999 } }).then(res => {
    if (res.data.code === 200) allCategories.value = res.data.data.list
  })
}

function saveCategory() {
  if (!catForm.name) { showToast('请输入分类名', 'error'); return }
  const fd = new URLSearchParams(); fd.append('name', catForm.name); fd.append('parent_id', catForm.parent_id)
  if (editingCat.value) {
    fd.append('identity', editingCat.value.identity)
    api.put('/admin/category-update', fd).then(res => {
      if (res.data.code === 200) { showToast('更新成功', 'success'); showCatForm.value = false; fetchCategories() }
      else showToast(res.data.msg, 'error')
    })
  } else {
    api.post('/admin/category-create', fd).then(res => {
      if (res.data.code === 200) { showToast('创建成功', 'success'); showCatForm.value = false; fetchCategories() }
      else showToast(res.data.msg, 'error')
    })
  }
}

function editCategory(c) {
  editingCat.value = c; catForm.name = c.name; catForm.parent_id = c.parent_id; showCatForm.value = true
}

function deleteCategory(identity) {
  if (!confirm('确定删除该分类？')) return
  api.delete('/admin/category-delete', { params: { identity } }).then(res => {
    if (res.data.code === 200) { showToast('删除成功', 'success'); fetchCategories() }
    else showToast(res.data.msg, 'error')
  })
}

// ======================== 题目管理 ========================
const adminTab = ref('categories')
const problemForm = reactive({
  title: '', content: '', max_runtime: 1000, max_mem: 256,
  category_ids: [], test_cases: [{ input: '', output: '' }]
})
const problemSaving = ref(false)
const problemFormMsg = ref('')
const problemFormError = ref(false)
const editingProblem = ref(null)
const adminProblems = ref([])
const problemSearch2 = reactive({ keyword: '' })

function resetProblemForm() {
  Object.assign(problemForm, { title: '', content: '', max_runtime: 1000, max_mem: 256, category_ids: [], test_cases: [{ input: '', output: '' }] })
  problemFormMsg.value = ''
}

function saveProblem() {
  const f = problemForm
  if (!f.title || !f.content || !f.category_ids.length || !f.test_cases.length || !f.max_runtime || !f.max_mem) {
    problemFormMsg.value = '请填写完整信息'; problemFormError.value = true; return
  }
  problemSaving.value = true; problemFormMsg.value = ''; problemFormError.value = false
  const fd = new URLSearchParams()
  fd.append('title', f.title); fd.append('content', f.content)
  fd.append('max_runtime', f.max_runtime); fd.append('max_mem', f.max_mem)
  f.category_ids.forEach(id => fd.append(editingProblem.value ? 'category_id' : 'category_ids', id))
  f.test_cases.forEach(tc => fd.append('test_cases', JSON.stringify({ input: tc.input, output: tc.output })))

  const req = editingProblem.value
    ? api.put('/admin/problem-update?' + new URLSearchParams({ identity: editingProblem.value.identity }), fd)
    : api.post('/admin/problem-create', fd)

  req.then(res => {
    const d = res.data
    if (d.code === 200) {
      showToast(editingProblem.value ? '更新成功' : '创建成功', 'success')
      editingProblem.value = null; resetProblemForm(); fetchAdminProblems()
    } else {
      problemFormMsg.value = d.msg || '操作失败'; problemFormError.value = true
    }
  }).catch(() => { problemFormMsg.value = '请求失败'; problemFormError.value = true }).finally(() => { problemSaving.value = false })
}

function fetchAdminProblems() {
  api.get('/problem-list', { params: { page: 1, size: 50, keyword: problemSearch2.keyword } }).then(res => {
    if (res.data.code === 200) adminProblems.value = res.data.data.list
  })
}

function loadProblemForEdit(identity) {
  adminTab.value = 'problem-create'
  api.get('/problem-detail', { params: { identity } }).then(res => {
    const d = res.data
    if (d.code === 200) {
      const p = d.data
      editingProblem.value = { identity: p.identity }
      problemForm.title = p.title; problemForm.content = p.content
      problemForm.max_runtime = p.max_runtime; problemForm.max_mem = p.max_mem
      problemForm.category_ids = (p.problem_categories || []).map(pc => String(pc.category_id))
      problemForm.test_cases = (p.test_cases || []).map(tc => ({ input: tc.input, output: tc.output }))
      problemFormMsg.value = ''
    }
  })
}

// ======================== 比赛列表 ========================
const contests = ref([])
const contestTotal = ref(0)
const contestPage = ref(1)
const contestSize = ref(20)
const contestLoading = ref(false)
const contestError = ref('')

async function fetchContests() {
  contestLoading.value = true; contestError.value = ''
  try {
    const res = await api.get('/contest-list', { params: { page: contestPage.value, size: contestSize.value } })
    if (res.data.code === 200) { contests.value = res.data.data.list; contestTotal.value = res.data.data.count }
    else contestError.value = res.data.msg
  } catch (e) { contestError.value = '网络错误' }
  finally { contestLoading.value = false }
}

function contestStatusClass(s) {
  return { 0: 'status-waiting', 1: 'status-running', 2: 'status-ended' }[s] || ''
}
function contestStatusText(s) {
  return { 0: '未开始', 1: '进行中', 2: '已结束' }[s] || '未知'
}

// ======================== 比赛详情 ========================
const contestDetail = ref(null)
const contestDetailLoading = ref(false)
const contestDetailError = ref('')
const contestMyScores = ref({})
const selectedContestProblem = ref(null)
const contestCode = ref('')
const contestSubmitting = ref(false)
const contestSubmitMsg = ref('')
const contestSubmitColor = ref('')
const contestSubmits = ref([])
const contestTab = ref('submit')
const contestRanks = ref([])
const contestRankLoading = ref(false)
const contestJoining = ref(false)
const contestJoinMsg = ref('')

async function openContestDetail(c) {
  currentView.value = 'contest-detail'
  contestDetailLoading.value = true; contestDetailError.value = ''
  contestMyScores.value = {}; selectedContestProblem.value = null
  contestSubmits.value = []; contestRanks.value = []; contestTab.value = 'submit'
  try {
    const res = await api.get('/contest-detail', { params: { identity: c.identity } })
    if (res.data.code === 200) {
      contestDetail.value = res.data.data
      const subRes = await api.get('/user/contest-submits', { params: { contest_identity: c.identity, size: 1000 } })
      if (subRes.data.code === 200) {
        const scores = {}
        for (const s of subRes.data.data.list) {
          const cur = scores[s.problem_identity] ?? 0
          if (s.score > cur) scores[s.problem_identity] = s.score
        }
        contestMyScores.value = scores
        contestSubmits.value = subRes.data.data.list
      }
    } else contestDetailError.value = res.data.msg
  } catch (e) { contestDetailError.value = '网络错误' }
  finally { contestDetailLoading.value = false }
}

function openContestProblem(cp) {
  selectedContestProblem.value = cp; contestCode.value = ''; contestSubmitMsg.value = ''
}

async function joinContest() {
  if (!contestDetail.value) return
  contestJoining.value = true; contestJoinMsg.value = ''
  try {
    const res = await api.post('/user/contest-join', null, { params: { contest_identity: contestDetail.value.contest.identity } })
    contestJoinMsg.value = res.data.code === 200 ? '报名成功' : res.data.msg
  } catch (e) { contestJoinMsg.value = '网络错误' }
  finally { contestJoining.value = false }
}

async function submitContestCode() {
  if (!selectedContestProblem.value || !contestCode.value.trim()) return
  contestSubmitting.value = true; contestSubmitMsg.value = ''
  try {
    const res = await api.post('/user/contest-submit', contestCode.value, {
      params: { contest_identity: contestDetail.value.contest.identity, problem_identity: selectedContestProblem.value.problem_identity },
      headers: { 'Content-Type': 'text/plain' }
    })
    if (res.data.code === 200) {
      contestSubmitMsg.value = res.data.data.msg + ' (得分: ' + res.data.data.score + ')'
      contestSubmitColor.value = res.data.data.status === 1 ? 'var(--success)' : 'var(--error)'
      const subRes = await api.get('/user/contest-submits', { params: { contest_identity: contestDetail.value.contest.identity, size: 1000 } })
      if (subRes.data.code === 200) {
        contestSubmits.value = subRes.data.data.list
        const scores = {}
        for (const s of subRes.data.data.list) {
          const cur = scores[s.problem_identity] ?? 0
          if (s.score > cur) scores[s.problem_identity] = s.score
        }
        contestMyScores.value = scores
      }
    } else {
      contestSubmitMsg.value = res.data.msg; contestSubmitColor.value = 'var(--error)'
    }
  } catch (e) { contestSubmitMsg.value = '网络错误'; contestSubmitColor.value = 'var(--error)' }
  finally { contestSubmitting.value = false }
}

async function fetchContestRank() {
  if (!contestDetail.value) return
  contestRankLoading.value = true
  try {
    const res = await api.get('/contest-rank', { params: { contest_identity: contestDetail.value.contest.identity, size: 100 } })
    if (res.data.code === 200) contestRanks.value = res.data.data.list
  } catch (e) { /* ignore */ }
  finally { contestRankLoading.value = false }
}

// ======================== 管理员比赛管理 ========================
const adminContests = ref([])
const adminContestLoading = ref(false)
const showContestForm = ref(false)
const editingContest = ref(null)
const contestForm = reactive({
  title: '', description: '', start_at: '', end_at: '', contest_type: 1, max_participants: 0, problem_identities: []
})
const contestSaving = ref(false)
const contestFormMsg = ref('')
const contestFormError = ref(false)
const allProblemsForContest = ref([])

function resetContestForm() {
  Object.assign(contestForm, { title: '', description: '', start_at: '', end_at: '', contest_type: 1, max_participants: 0, problem_identities: [] })
  contestFormMsg.value = ''
}

function formatDateTimeLocal(dt) {
  if (!dt) return ''
  const d = new Date(dt)
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

async function fetchAdminContests() {
  adminContestLoading.value = true
  try {
    const res = await api.get('/admin/contest-list', { params: { page: 1, size: 50 } })
    if (res.data.code === 200) adminContests.value = res.data.data.list
  } catch (e) { /* ignore */ }
  finally { adminContestLoading.value = false }
  // 同时加载题目列表供选择
  api.get('/problem-list', { params: { page: 1, size: 200 } }).then(res => {
    if (res.data.code === 200) allProblemsForContest.value = res.data.data.list
  }).catch(() => {})
}

async function saveContest() {
  const f = contestForm
  if (!f.title || !f.start_at || !f.end_at || !f.problem_identities.length) {
    contestFormMsg.value = '请填写完整信息'; contestFormError.value = true; return
  }
  contestSaving.value = true; contestFormMsg.value = ''; contestFormError.value = false
  const fd = new URLSearchParams()
  fd.append('title', f.title); fd.append('description', f.description)
  fd.append('start_at', formatDateTimeLocal(f.start_at)); fd.append('end_at', formatDateTimeLocal(f.end_at))
  fd.append('contest_type', f.contest_type); fd.append('max_participants', f.max_participants)
  fd.append('problem_identities', JSON.stringify(f.problem_identities))

  const req = editingContest.value
    ? api.put('/admin/contest-update?' + new URLSearchParams({ identity: editingContest.value.identity }), fd)
    : api.post('/admin/contest-create', fd)

  try {
    const res = await req
    if (res.data.code === 200) {
      showToast(editingContest.value ? '更新成功' : '创建成功', 'success')
      editingContest.value = null; showContestForm.value = false; resetContestForm(); fetchAdminContests()
    } else {
      contestFormMsg.value = res.data.msg || '操作失败'; contestFormError.value = true
    }
  } catch (e) { contestFormMsg.value = '请求失败'; contestFormError.value = true }
  finally { contestSaving.value = false }
}

async function loadContestForEdit(c) {
  showContestForm.value = true
  try {
    const res = await api.get('/contest-detail', { params: { identity: c.identity } })
    if (res.data.code === 200) {
      const d = res.data.data.contest
      editingContest.value = { identity: d.identity }
      contestForm.title = d.title; contestForm.description = d.description
      contestForm.start_at = d.start_at?.replace(' ', 'T').slice(0, 16) || ''
      contestForm.end_at = d.end_at?.replace(' ', 'T').slice(0, 16) || ''
      contestForm.contest_type = d.contest_type; contestForm.max_participants = d.max_participants || 0
      contestForm.problem_identities = (d.contest_problems || []).map(cp => cp.problem_identity)
      contestFormMsg.value = ''
    }
  } catch (e) { /* ignore */ }
}

async function deleteContest(identity) {
  if (!confirm('确定删除该比赛？')) return
  try {
    const res = await api.delete('/admin/contest-delete', { params: { identity } })
    if (res.data.code === 200) { showToast('删除成功', 'success'); fetchAdminContests() }
    else showToast(res.data.msg, 'error')
  } catch (e) { showToast('删除失败', 'error') }
}

// ======================== 初始化 ========================
onMounted(() => {
  fetchProblems()
  // 预加载分类列表（仅管理员可用，普通用户静默失败）
  if (isAdmin.value) {
    api.get('/admin/category-list', { params: { page: 1, size: 999 } }).then(res => {
      if (res.data.code === 200) categories.value = res.data.data.list
    }).catch(() => {})
  }
})
</script>

<style>
/* ======================== CSS Reset & Variables ======================== */
* { margin: 0; padding: 0; box-sizing: border-box; }

:root {
  --primary: #4f46e5;
  --primary-hover: #4338ca;
  --primary-light: #eef2ff;
  --success: #059669;
  --success-bg: #ecfdf5;
  --error: #dc2626;
  --error-bg: #fef2f2;
  --warn: #d97706;
  --warn-bg: #fffbeb;
  --bg: #f3f4f6;
  --surface: #ffffff;
  --border: #e5e7eb;
  --text: #111827;
  --text-secondary: #6b7280;
  --radius: 8px;
  --shadow: 0 1px 3px rgba(0,0,0,.08), 0 1px 2px rgba(0,0,0,.06);
  --shadow-lg: 0 10px 25px rgba(0,0,0,.1);
  --font: 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

body { background: var(--bg); font-family: var(--font); color: var(--text); -webkit-font-smoothing: antialiased; }

/* ======================== Layout ======================== */
.oj-header {
  background: var(--surface); border-bottom: 1px solid var(--border);
  position: sticky; top: 0; z-index: 100; box-shadow: var(--shadow);
}
.header-inner {
  max-width: 1200px; margin: 0 auto; padding: 0 24px;
  display: flex; align-items: center; height: 60px; gap: 32px;
}
.logo {
  display: flex; align-items: center; gap: 8px; font-size: 18px; font-weight: 700;
  color: var(--primary); cursor: pointer; user-select: none; white-space: nowrap;
}
.logo-icon { width: 24px; height: 24px; }
.header-nav { display: flex; gap: 4px; flex: 1; }
.header-nav a {
  padding: 8px 16px; border-radius: var(--radius); font-size: 14px; font-weight: 500;
  color: var(--text-secondary); cursor: pointer; transition: all .2s; text-decoration: none;
}
.header-nav a:hover { background: var(--primary-light); color: var(--primary); }
.header-nav a.active { background: var(--primary-light); color: var(--primary); font-weight: 600; }
.header-right { display: flex; align-items: center; gap: 12px; white-space: nowrap; }
.user-name { font-size: 14px; color: var(--text-secondary); }

.container { max-width: 1200px; margin: 0 auto; padding: 24px; }
.page-header { margin-bottom: 24px; }
.page-header h2 { font-size: 24px; font-weight: 700; }
.subtitle { color: var(--text-secondary); margin-top: 4px; font-size: 14px; }

/* ======================== Buttons ======================== */
.btn {
  padding: 8px 20px; border: none; border-radius: var(--radius); font-size: 14px; font-weight: 500;
  cursor: pointer; transition: all .2s; display: inline-flex; align-items: center; justify-content: center;
  gap: 6px; font-family: var(--font);
}
.btn:disabled { opacity: .5; cursor: not-allowed; }
.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover:not(:disabled) { background: var(--primary-hover); }
.btn-outline { background: transparent; color: var(--text); border: 1px solid var(--border); }
.btn-outline:hover:not(:disabled) { background: var(--bg); border-color: #9ca3af; }
.btn-success { background: var(--success); color: #fff; }
.btn-success:hover:not(:disabled) { background: #047857; }
.btn-danger { background: var(--error); color: #fff; }
.btn-danger:hover:not(:disabled) { background: #b91c1c; }
.btn-text { background: none; border: none; color: var(--primary); padding: 8px 0; font-size: 14px; }
.btn-text:hover { text-decoration: underline; }
.btn-sm { padding: 4px 12px; font-size: 12px; border-radius: 6px; }
.btn-block { width: 100%; }
.btn-close { background: none; border: none; font-size: 22px; color: var(--text-secondary); cursor: pointer; line-height: 1; }

/* ======================== Input ======================== */
.input {
  width: 100%; padding: 10px 14px; border: 1px solid var(--border); border-radius: var(--radius);
  font-size: 14px; font-family: var(--font); transition: border .2s; background: var(--surface);
  color: var(--text);
}
.input:focus { outline: none; border-color: var(--primary); box-shadow: 0 0 0 3px rgba(79,70,229,.1); }
.select { cursor: pointer; appearance: auto; min-width: 150px; }
.textarea { resize: vertical; min-height: 120px; }

/* ======================== Data Table ======================== */
.data-table { width: 100%; border-collapse: collapse; background: var(--surface); border-radius: var(--radius); overflow: hidden; box-shadow: var(--shadow); }
.data-table th { background: #f9fafb; padding: 12px 16px; text-align: left; font-weight: 600; font-size: 13px; color: var(--text-secondary); border-bottom: 1px solid var(--border); }
.data-table td { padding: 12px 16px; border-bottom: 1px solid var(--border); font-size: 14px; }
.data-table tr:last-child td { border-bottom: none; }
.data-table tr.clickable { cursor: pointer; transition: background .15s; }
.data-table tr.clickable:hover { background: var(--primary-light); }
.status-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; background: var(--border); }
.status-dot.passed { background: var(--success); }
.status-tag {
  display: inline-block; padding: 2px 10px; border-radius: 12px; font-size: 12px; font-weight: 600;
}
.status-tag.success { background: var(--success-bg); color: var(--success); }
.status-tag.error { background: var(--error-bg); color: var(--error); }
.status-tag.warn { background: var(--warn-bg); color: var(--warn); }
.rank-badge {
  display: inline-flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 50%; font-size: 13px; font-weight: 700;
  background: var(--bg); color: var(--text-secondary);
}
.rank-badge.rank-1 { background: #fef3c7; color: #b45309; }
.rank-badge.rank-2 { background: #e5e7eb; color: #4b5563; }
.rank-badge.rank-3 { background: #fed7aa; color: #9a3412; }

/* ======================== Pagination ======================== */
.pagination {
  display: flex; align-items: center; justify-content: center; gap: 16px; padding: 16px 0;
  font-size: 14px; color: var(--text-secondary);
}
.pagination button {
  padding: 6px 16px; border: 1px solid var(--border); border-radius: var(--radius);
  background: var(--surface); cursor: pointer; font-size: 14px; transition: all .2s;
}
.pagination button:hover:not(:disabled) { border-color: var(--primary); color: var(--primary); }
.pagination button:disabled { opacity: .4; cursor: not-allowed; }

/* ======================== Search Bar ======================== */
.search-bar { display: flex; gap: 12px; margin-bottom: 20px; flex-wrap: wrap; }
.search-bar .input { width: auto; flex: 1; min-width: 160px; }
.search-bar .select { width: 180px; flex: none; }
.search-bar .btn { flex: none; }

/* ======================== State Boxes ======================== */
.state-box {
  text-align: center; padding: 48px 24px; color: var(--text-secondary); font-size: 15px;
}
.state-box.error { color: var(--error); }
.state-box.empty { color: var(--text-secondary); }
.spinner {
  display: inline-block; width: 20px; height: 20px; border: 2px solid var(--border);
  border-top-color: var(--primary); border-radius: 50%; animation: spin .6s linear infinite;
  vertical-align: middle; margin-right: 8px;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ======================== Detail Page ======================== */
.detail-layout { display: grid; grid-template-columns: 1fr 400px; gap: 24px; align-items: start; }
@media (max-width: 900px) { .detail-layout { grid-template-columns: 1fr; } }
.detail-left {
  background: var(--surface); border-radius: var(--radius); padding: 24px; box-shadow: var(--shadow);
}
.detail-left h2 { font-size: 22px; margin-bottom: 12px; }
.detail-meta { display: flex; flex-wrap: wrap; gap: 16px; margin-bottom: 20px; padding-bottom: 16px; border-bottom: 1px solid var(--border); }
.detail-meta span { font-size: 13px; color: var(--text-secondary); }
.detail-content { font-size: 15px; line-height: 1.8; }
.detail-content pre { background: #1e293b; color: #e2e8f0; padding: 16px; border-radius: var(--radius); overflow-x: auto; }
.detail-right { position: sticky; top: 80px; }
.editor-panel { background: var(--surface); border-radius: var(--radius); padding: 20px; box-shadow: var(--shadow); }
.panel-header { font-weight: 600; margin-bottom: 12px; font-size: 15px; }
.code-editor {
  width: 100%; height: 320px; padding: 16px; font-family: 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
  font-size: 13px; line-height: 1.6; border: 1px solid var(--border); border-radius: var(--radius);
  background: #1e293b; color: #e2e8f0; resize: vertical;
}
.code-editor:focus { outline: none; border-color: var(--primary); }
.code-editor::placeholder { color: #64748b; }
.submit-result { margin-top: 12px; padding: 10px 14px; border-radius: var(--radius); font-size: 14px; font-weight: 600; text-align: center; }
.submit-result.success { background: var(--success-bg); color: var(--success); }
.submit-result.error { background: var(--error-bg); color: var(--error); }

/* ======================== Admin ======================== */
.admin-tabs { display: flex; gap: 4px; margin-bottom: 20px; }
.admin-tabs button {
  padding: 8px 20px; border: 1px solid var(--border); border-radius: var(--radius);
  background: var(--surface); cursor: pointer; font-size: 14px; transition: all .2s;
}
.admin-tabs button.active { background: var(--primary); color: #fff; border-color: var(--primary); }
.admin-section { margin-bottom: 24px; }
.form-card {
  background: var(--surface); border-radius: var(--radius); padding: 24px;
  box-shadow: var(--shadow); display: flex; flex-direction: column; gap: 14px;
  margin-bottom: 20px;
}
.form-card h4 { font-size: 16px; margin-bottom: 4px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.form-group label { display: block; font-size: 14px; font-weight: 500; margin-bottom: 6px; }
.checkbox-group { display: flex; flex-wrap: wrap; gap: 8px; }
.checkbox-label {
  display: flex; align-items: center; gap: 4px; font-size: 14px; cursor: pointer;
  padding: 4px 10px; border: 1px solid var(--border); border-radius: var(--radius);
  transition: all .2s;
}
.checkbox-label:has(input:checked) { background: var(--primary-light); border-color: var(--primary); color: var(--primary); }
.test-case-row { display: flex; gap: 8px; align-items: center; }
.test-case-row .input { flex: 1; }
.form-actions { display: flex; gap: 12px; align-items: center; }
.form-msg { font-size: 13px; color: var(--success); }
.form-msg.error { color: var(--error); }

/* ======================== Toast ======================== */
.toast {
  position: fixed; top: 24px; left: 50%; transform: translateX(-50%);
  padding: 12px 24px; border-radius: var(--radius); font-size: 14px; font-weight: 500;
  z-index: 999; box-shadow: var(--shadow-lg);
}
.toast.info { background: #1e293b; color: #fff; }
.toast.success { background: var(--success); color: #fff; }
.toast.error { background: var(--error); color: #fff; }
.toast-fade-enter-active { transition: all .3s ease-out; }
.toast-fade-leave-active { transition: all .25s ease-in; }
.toast-fade-enter-from, .toast-fade-leave-to { opacity: 0; transform: translateX(-50%) translateY(-12px); }

/* ======================== Contest ======================== */
.status-badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
.status-waiting { background: #e6f7ff; color: #1890ff; }
.status-running { background: #f6ffed; color: #52c41a; }
.status-ended { background: #fff1f0; color: #ff4d4f; }
.right-tabs { display: flex; gap: 12px; margin-bottom: 12px; border-bottom: 1px solid var(--border); padding-bottom: 8px; }
.right-tabs a { cursor: pointer; padding: 4px 8px; font-size: 14px; color: var(--text-secondary); transition: all .2s; }
.right-tabs a.active { border-bottom: 2px solid var(--primary); color: var(--primary); font-weight: 600; }
.selected-row { background: var(--primary-light) !important; }
</style>
