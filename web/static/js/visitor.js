const escapeHTML = (value) => String(value).replace(/[&<>"']/g, (char) => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
}[char]));

const formatFileSize = (bytes) => {
    const size = Number(bytes) || 0;
    if (size < 1024) return `${size} B`;
    if (size < 1024 ** 2) return `${(size / 1024).toFixed(1)} KB`;
    if (size < 1024 ** 3) return `${(size / 1024 ** 2).toFixed(1)} MB`;
    return `${(size / 1024 ** 3).toFixed(1)} GB`;
};

const formatUpdateTime = (value) => {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '更新时间未知';
    return `更新于 ${new Intl.DateTimeFormat('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    }).format(date)}`;
};

const getFileCategory = (fileName, isDir) => {
    if (isDir) return 'folder';
    const extension = String(fileName).split('.').pop().toLowerCase();
    if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'heic', 'bmp'].includes(extension)) return 'image';
    if (['zip', 'rar', '7z', 'tar', 'gz', 'bz2'].includes(extension)) return 'archive';
    if (['mp3', 'wav', 'flac', 'aac', 'm4a', 'ogg'].includes(extension)) return 'audio';
    if (['mp4', 'mov', 'mkv', 'avi', 'webm', 'm4v'].includes(extension)) return 'video';
    if (['js', 'ts', 'jsx', 'tsx', 'html', 'css', 'go', 'py', 'java', 'json', 'yml', 'yaml', 'md'].includes(extension)) return 'code';
    return 'document';
};

const getFileIcon = (category) => {
    const icons = {
        folder: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3.5 6.5h7l2 2.5h8v9.5h-17z"></path><path d="M3.5 9h17"></path></svg>',
        image: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="4" y="4" width="16" height="16" rx="2"></rect><circle cx="9" cy="9" r="1.5"></circle><path d="m6 17 4.2-4.5 3.1 3 2.2-2.1L19 17"></path></svg>',
        archive: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 4h12v16H6z"></path><path d="M9 4v3h3V4m-3 6h3v3H9m0 3h3v4"></path></svg>',
        audio: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M9 17.5V7l9-2v10.5"></path><circle cx="6.5" cy="17.5" r="2.5"></circle><circle cx="15.5" cy="15.5" r="2.5"></circle></svg>',
        video: '<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3.5" y="5" width="17" height="14" rx="2"></rect><path d="m10 9 5 3-5 3z"></path></svg>',
        code: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m9 7-5 5 5 5m6-10 5 5-5 5m-2-12-2 14"></path></svg>',
        document: '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 3.5h8l4 4V20H6z"></path><path d="M14 3.5v4h4M9 12h6m-6 3h6"></path></svg>'
    };
    return icons[category] || icons.document;
};

const getKindLabel = (fileName, isDir) => {
    if (isDir) return '文件夹';
    const parts = String(fileName).split('.');
    return parts.length > 1 ? parts.pop().toUpperCase() : '文件';
};

let allFiles = [];

function updateResultCount(count) {
    const resultCount = document.getElementById('result-count');
    if (resultCount) resultCount.textContent = String(count);
}

async function loadFiles() {
    const listEl = document.getElementById('file-list');
    listEl.setAttribute('aria-busy', 'true');

    try {
        const response = await fetch('/api/files');
        if (!response.ok) throw new Error('加载文件列表失败');
        const result = await response.json();
        allFiles = Array.isArray(result) ? result : [];
    } catch (error) {
        allFiles = [];
        listEl.innerHTML = '<li class="empty-message error-message">文件列表暂时无法加载，请稍后刷新</li>';
        updateResultCount(0);
        document.getElementById('archive-count').textContent = '0';
        listEl.setAttribute('aria-busy', 'false');
        return;
    }

    document.getElementById('archive-count').textContent = String(allFiles.length);
    renderFiles(allFiles, '这里还没有公开文件');
    listEl.setAttribute('aria-busy', 'false');
}

function renderFiles(files, emptyText = '没有找到匹配的文件') {
    const listEl = document.getElementById('file-list');
    updateResultCount(files.length);

    if (files.length === 0) {
        listEl.innerHTML = `<li class="empty-message">${emptyText}</li>`;
        return;
    }

    listEl.innerHTML = files.map((file) => {
        const isDir = Boolean(file.is_dir);
        const safeFileName = escapeHTML(file.file_name);
        const category = getFileCategory(file.file_name, isDir);
        const kindText = getKindLabel(file.file_name, isDir);
        const downloadName = isDir ? `${safeFileName}.zip` : safeFileName;
        const downloadURL = `/api/download/${encodeURIComponent(file.id)}`;

        return `
            <li class="file-item${isDir ? ' folder-item' : ''}">
                <span class="file-icon file-icon-${category}" aria-hidden="true">${getFileIcon(category)}</span>
                <div class="file-info">
                    <div class="file-name" title="${safeFileName}">${safeFileName}</div>
                    <div class="file-meta">
                        <span class="file-kind">${kindText}</span>
                        <span class="file-size">${formatFileSize(file.file_size)}</span>
                        <span class="file-time">${formatUpdateTime(file.update_time)}</span>
                    </div>
                </div>
                <div class="file-actions">
                    <a href="${downloadURL}" class="download-btn" download="${downloadName}" aria-label="下载 ${safeFileName}">
                        <svg viewBox="0 0 18 18" aria-hidden="true"><path d="M9 2.5v9m-3.5-3L9 12l3.5-3.5M3 15h12"></path></svg>
                        下载
                    </a>
                </div>
            </li>
        `;
    }).join('');
}

function searchFiles() {
    const searchTerm = document.getElementById('search-input').value.toLowerCase().trim();
    if (!searchTerm) {
        renderFiles(allFiles, '这里还没有公开文件');
        return;
    }

    const filteredFiles = allFiles.filter((file) =>
        String(file.file_name).toLowerCase().includes(searchTerm)
    );
    renderFiles(filteredFiles);
}

function resetSearch() {
    const input = document.getElementById('search-input');
    input.value = '';
    renderFiles(allFiles, '这里还没有公开文件');
    input.focus();
}

function initSearch() {
    const input = document.getElementById('search-input');
    document.getElementById('search-btn').addEventListener('click', searchFiles);
    document.getElementById('reset-btn').addEventListener('click', resetSearch);
    input.addEventListener('input', searchFiles);
    input.addEventListener('keydown', (event) => {
        if (event.key === 'Escape') resetSearch();
    });
}

document.getElementById('current-year').textContent = String(new Date().getFullYear());
loadFiles();
initSearch();
