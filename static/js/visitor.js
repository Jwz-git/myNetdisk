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

// 初始化
loadFiles();