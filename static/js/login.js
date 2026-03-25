document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('login-form');
    const errorElement = document.getElementById('login-error');
    
    if (loginForm) {
        loginForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            
            // 清空之前的错误信息
            errorElement.textContent = '';
            
            // 简单的表单验证
            if (!username || !password) {
                errorElement.textContent = '请输入用户名和密码';
                return;
            }
            
            // 发送登录请求
            fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password })
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error('网络响应错误');
                }
                return response.json();
            })
            .then(data => {
                if (data.success) {
                    window.location.href = '/admin';
                } else {
                    errorElement.textContent = data.error || '登录失败';
                }
            })
            .catch(error => {
                console.error('登录错误:', error);
                errorElement.textContent = '登录过程中发生错误';
            });
        });
    }
});