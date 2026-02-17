// 搜索功能
let allFiles = [];

// 加载文件列表并保存到全局变量
async function loadFiles() {
    const res = await fetch('/api/files');
    allFiles = await res.json();
    const listEl = document.getElementById('file-list');
    if (allFiles.length === 0) {
        listEl.innerHTML = '<li class="empty-message">暂无文件</li>';
        return;
    }

    renderFiles(allFiles);
}

// 渲染文件列表
function renderFiles(files) {
    const listEl = document.getElementById('file-list');
    if (files.length === 0) {
        listEl.innerHTML = '<li class="empty-message">没有找到匹配的文件</li>';
        return;
    }

    const formatFileSize = (bytes) => {
        if (bytes < 1024) {
            return bytes + ' B';
        } else if (bytes < 1024 * 1024) {
            return (bytes / 1024).toFixed(2) + ' KB';
        } else if (bytes < 1024 * 1024 * 1024) {
            return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
        }
        return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
    };

    listEl.innerHTML = files.map(f => `
        <li class="file-item">
            <div class="file-info">
                <div class="file-name">${f.file_name}</div>
                <div class="file-meta">
                    <span class="file-size">${formatFileSize(f.file_size)}</span>
                    <span class="file-time">更新时间：${new Date(f.update_time).toLocaleString()}</span>
                </div>
            </div>
            <div class="file-actions">
                <a href="/api/download/${f.id}" class="download-btn" download="${f.file_name}">下载</a>
            </div>
        </li>
    `).join('');
}

// 搜索文件
function searchFiles() {
    const searchTerm = document.getElementById('search-input').value.toLowerCase().trim();
    if (!searchTerm) {
        renderFiles(allFiles);
        return;
    }
    const filteredFiles = allFiles.filter(file => 
        file.file_name.toLowerCase().includes(searchTerm)
    );
    renderFiles(filteredFiles);
}

// 重置搜索
function resetSearch() {
    document.getElementById('search-input').value = '';
    renderFiles(allFiles);
}

// 初始化搜索事件监听器
function initSearch() {
    document.getElementById('search-btn').addEventListener('click', searchFiles);
    document.getElementById('reset-btn').addEventListener('click', resetSearch);
    // 支持回车键搜索
    document.getElementById('search-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            searchFiles();
        }
    });
}

// 初始化
loadFiles();
initSearch();