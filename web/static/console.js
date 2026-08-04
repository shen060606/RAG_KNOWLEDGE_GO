const pageTitles = {
    overview: '账户概览',
    profile: '修改用户名',
    security: '账号安全'
};

let toastTimer;

async function apiRequest(url, options = {}) {
    const response = await fetch(url, {
        credentials: 'same-origin',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...(options.headers || {})
        }
    });

    if (response.status === 401) {
        window.location.href = '/login';
        return null;
    }

    let data = {};
    try {
        data = await response.json();
    } catch (_) {
        data = {};
    }

    if (!response.ok) {
        throw new Error(data.msg || '请求失败，请稍后重试');
    }

    return data;
}

function showToast(message, isError = false) {
    const toast = document.getElementById('toast');
    toast.textContent = message;
    toast.classList.toggle('error', isError);
    toast.classList.add('show');

    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => toast.classList.remove('show'), 2600);
}

function formatDate(value) {
    if (!value) return '-';

    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '-';

    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function renderProfile(user) {
    const roleName = user.role === 'admin' ? '管理员' : '普通用户';
    const avatarText = Array.from(user.username || 'U')[0].toUpperCase();

    document.getElementById('top-username').textContent = user.username;
    document.getElementById('top-role').textContent = roleName;
    document.getElementById('account-avatar').textContent = avatarText;
    document.getElementById('role-badge').textContent = roleName;
    document.getElementById('welcome-username').textContent = user.username;
    document.getElementById('detail-user-id').textContent = user.userID;
    document.getElementById('detail-username').textContent = user.username;
    document.getElementById('detail-role').textContent = roleName;
    document.getElementById('detail-created-at').textContent = formatDate(user.createdAt);
    document.getElementById('new-username').value = user.username;
}

function renderStatistics(stats) {
    document.getElementById('document-count').textContent = stats.document_count ?? 0;
    document.getElementById('chunk-count').textContent = stats.chunk_count ?? 0;
    document.getElementById('question-count').textContent = stats.question_count ?? 0;
}

async function loadConsole() {
    try {
        const [user, stats] = await Promise.all([
            apiRequest('/api/user/me'),
            apiRequest('/api/user/statistics')
        ]);

        if (!user || !stats) return;

        renderProfile(user);
        renderStatistics(stats);
        document.getElementById('page-loading').hidden = true;
        document.getElementById('console-content').hidden = false;
    } catch (error) {
        document.getElementById('page-loading').textContent = '用户中心加载失败';
        showToast(error.message, true);
    }
}

function switchSection(sectionName) {
    for (const section of document.querySelectorAll('.page-section')) {
        section.hidden = section.id !== `${sectionName}-section`;
    }

    for (const item of document.querySelectorAll('.nav-item')) {
        item.classList.toggle('active', item.dataset.section === sectionName);
    }

    document.getElementById('page-title').textContent = pageTitles[sectionName];
}

for (const item of document.querySelectorAll('.nav-item')) {
    item.addEventListener('click', () => switchSection(item.dataset.section));
}

document.getElementById('username-form').addEventListener('submit', async event => {
    event.preventDefault();

    const button = document.getElementById('username-submit');
    const username = document.getElementById('new-username').value.trim();

    button.disabled = true;
    button.textContent = '保存中...';

    try {
        const data = await apiRequest('/api/user/username', {
            method: 'PATCH',
            body: JSON.stringify({ username })
        });

        if (!data) return;

        document.getElementById('top-username').textContent = data.username;
        document.getElementById('welcome-username').textContent = data.username;
        document.getElementById('detail-username').textContent = data.username;
        document.getElementById('account-avatar').textContent =
            Array.from(data.username || 'U')[0].toUpperCase();
        showToast('用户名修改成功');
    } catch (error) {
        showToast(error.message, true);
    } finally {
        button.disabled = false;
        button.textContent = '保存修改';
    }
});

document.getElementById('password-form').addEventListener('submit', async event => {
    event.preventDefault();

    const oldPassword = document.getElementById('old-password').value;
    const newPassword = document.getElementById('new-password').value;
    const confirmPassword = document.getElementById('confirm-password').value;

    if (newPassword !== confirmPassword) {
        showToast('两次输入的新密码不一致', true);
        return;
    }

    const button = document.getElementById('password-submit');
    button.disabled = true;
    button.textContent = '更新中...';

    try {
        const data = await apiRequest('/api/user/password', {
            method: 'PATCH',
            body: JSON.stringify({
                oldPassword,
                newPassword,
                confirmPassword
            })
        });

        if (!data) return;

        alert('密码修改成功，请使用新密码重新登录。');
        window.location.href = '/login';
    } catch (error) {
        showToast(error.message, true);
        button.disabled = false;
        button.textContent = '更新密码';
    }
});

document.getElementById('logout-button').addEventListener('click', async () => {
    try {
        await apiRequest('/api/logout', { method: 'POST' });
    } finally {
        window.location.href = '/login';
    }
});

loadConsole();
