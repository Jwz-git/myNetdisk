document.addEventListener('DOMContentLoaded', () => {
    const loginForm = document.getElementById('login-form');
    const errorElement = document.getElementById('login-error');
    const passwordInput = document.getElementById('password');
    const passwordToggle = document.getElementById('password-toggle');
    const submitButton = loginForm.querySelector('.login-btn');
    const defaultButtonContent = submitButton.innerHTML;

    passwordToggle.addEventListener('click', () => {
        const shouldShow = passwordInput.type === 'password';
        passwordInput.type = shouldShow ? 'text' : 'password';
        passwordToggle.textContent = shouldShow ? '隐藏' : '显示';
        passwordToggle.setAttribute('aria-label', shouldShow ? '隐藏密码' : '显示密码');
        passwordToggle.setAttribute('aria-pressed', String(shouldShow));
        passwordInput.focus();
    });

    loginForm.addEventListener('submit', async (event) => {
        event.preventDefault();

        const username = document.getElementById('username').value.trim();
        const password = passwordInput.value;
        errorElement.textContent = '';

        if (!username || !password) {
            errorElement.textContent = '请输入用户名和密码';
            return;
        }

        submitButton.disabled = true;
        submitButton.textContent = '正在验证身份…';

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password })
            });

            const data = response.headers.get('Content-Type')?.includes('application/json')
                ? await response.json()
                : { success: false, error: '服务暂时不可用' };

            if (!response.ok || !data.success) {
                throw new Error(data.error || '登录失败');
            }

            submitButton.textContent = '验证通过，正在进入…';
            window.location.href = '/admin';
        } catch (error) {
            errorElement.textContent = error.message || '登录过程中发生错误';
            submitButton.disabled = false;
            submitButton.innerHTML = defaultButtonContent;
        }
    });
});
