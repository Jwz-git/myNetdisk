// 加载文件列表
async function loadFiles() {
    const res = await fetch('/api/files');
    const files = await res.json();
    const listEl = document.getElementById('file-list');
    if (files.length === 0) {
        listEl.innerHTML = '<li class="empty-message">暂无文件</li>';
        return;
    }

        const formatFileSize = (bytes) => {
        if (bytes < 1024) {
            return bytes + ' B';
        } else if (bytes < 1024 * 1024) {
            return (bytes / 1024).toFixed(2) + ' KB';
        } else {
            return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
        }
    };

    listEl.innerHTML = files.map(f => `
        <li class="file-item">
            <div class="file-info">
                <div class="file-name">${f.file_name}</div>
                <div class="file-meta">
                    <span class="file-size">${formatFileSize(f.file_size)}</span>
                    <span class="file-time">更新时间：${new Date(f.upload_time).toLocaleString()}</span>
                </div>
            </div>
            <div class="file-actions">
                <a href="/api/download/${f.id}" class="download-btn" download>下载</a>
            </div>
        </li>
    `).join('');
}

// 上传文件
document.getElementById('upload-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    const statusEl = document.getElementById('upload-status');
    statusEl.textContent = '上传中...';
    try {
        const res = await fetch('/api/upload', { method: 'POST', body: formData });
        const result = await res.json();
        if (result.success) {
            statusEl.textContent = '上传成功！';
            loadFiles(); // 刷新列表
            e.target.reset(); // 清空选择
        } else {
            statusEl.textContent = '上传失败：' + result.error;
        }
    } catch (err) {
        statusEl.textContent = '网络错误';
    }
});

// 初始化
loadFiles();