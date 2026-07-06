const escapeHTML = (value) => String(value).replace(/[&<>"']/g, (char) => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
}[char]));

const getUploadPath = (file) => file.webkitRelativePath || file.name;

const deleteFile = async (id, isDir = false) => {
    const message = isDir ? '确定要删除这个文件夹及其中内容吗？' : '确定要删除这个文件吗？';
    if (!confirm(message)) return;

    try {
        const response = await fetch(`/api/delete/${id}`, { method: 'DELETE' });
        const data = response.headers.get('Content-Type')?.includes('application/json')
            ? await response.json()
            : { success: false, error: await response.text() };

        if (data.success) {
            loadFiles();
        } else {
            alert('删除失败：' + (data.error || '未知错误'));
        }
    } catch (error) {
        alert('网络错误，请稍后重试');
        console.error('删除失败:', error);
    }
}

// 上传文件
document.getElementById('upload-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    if (selectedFilesList.length === 0) {
        document.getElementById('upload-status').textContent = '请先选择文件';
        return;
    }
    
    const formData = new FormData();
    for (let i = 0; i < selectedFilesList.length; i++) {
        const file = selectedFilesList[i];
        formData.append('files', file);
        formData.append('paths', getUploadPath(file));
    }
    
    const statusEl = document.getElementById('upload-status');
    statusEl.textContent = '上传中...';
    try {
        const res = await fetch('/api/upload', { method: 'POST', body: formData });
        const result = await res.json();
        if (result.success) {
            statusEl.textContent = '上传成功！';
            loadFiles(); // 刷新列表
            selectedFilesList = []; // 清空选择
            showSelectedFiles(); // 更新显示
        } else {
            statusEl.textContent = '上传失败：' + result.error;
        }
    } catch (err) {
        statusEl.textContent = '网络错误';
    }

    loadFiles(); // 刷新列表
});

// 重命名功能
function startRename(id) {
    // 隐藏文件名和重命名按钮
    document.getElementById(`file-name-${id}`).style.display = 'none';
    document.querySelector(`#save-btn-${id}`).style.display = 'none';
    document.querySelector(`#cancel-btn-${id}`).style.display = 'none';
    document.getElementById(`edit-btn-${id}`).style.display = 'none';
    
    // 显示输入框和保存/取消按钮
    document.getElementById(`rename-container-${id}`).style.display = 'flex';
    document.querySelector(`#save-btn-${id}`).style.display = 'inline-block';
    document.querySelector(`#cancel-btn-${id}`).style.display = 'inline-block';
    
    // 聚焦输入框
    const input = document.getElementById(`rename-input-${id}`);
    input.focus();
    input.select();
}

function cancelRename(id) {
    // 隐藏输入框和保存/取消按钮
    document.getElementById(`rename-container-${id}`).style.display = 'none';
    document.querySelector(`#save-btn-${id}`).style.display = 'none';
    document.querySelector(`#cancel-btn-${id}`).style.display = 'none';
    
    // 显示文件名和重命名按钮
    document.getElementById(`file-name-${id}`).style.display = 'block';
    const editBtn = document.getElementById(`edit-btn-${id}`);
    if (editBtn) {
        editBtn.style.display = 'inline-block';
    }
}

async function saveRename(id) {
    const input = document.getElementById(`rename-input-${id}`);
    const newNameWithoutExt = input.value.trim();
    
    if (!newNameWithoutExt) {
        alert('文件名不能为空');
        return;
    }
    
    // 获取扩展名
    const extensionElement = document.querySelector(`#rename-container-${id} .file-extension`);
    const extension = extensionElement.textContent;
    
    const newFileName = newNameWithoutExt + extension;
    
    try {
        const response = await fetch(`/api/rename/${id}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ file_name: newFileName })
        });

        // console.log('重命名请求响应:', response);
        
        const result = await response.json();
        if (result.success) {
            // 更新文件名显示
            document.getElementById(`file-name-${id}`).textContent = newFileName;
            cancelRename(id);
        } else {
            alert('重命名失败：' + (result.error || '未知错误'));
        }
        // console.log('重命名结果:', result);

        loadFiles(); // 刷新列表
        
    } catch (error) {
        alert('网络错误，请稍后重试');
        console.error('重命名失败:', error);
    }
}

// 搜索功能
let allFiles = [];
// 存储所有选择的文件
let selectedFilesList = [];

// 加载文件列表并保存到全局变量
async function loadFiles() {
    const listEl = document.getElementById('file-list');
    try {
        const res = await fetch('/api/files');
        if (!res.ok) {
            throw new Error('加载文件列表失败');
        }
        allFiles = await res.json();
    } catch (error) {
        allFiles = [];
        listEl.innerHTML = '<li class="empty-message">暂无文件</li>';
        return;
    }

    if (allFiles.length === 0) {
        listEl.innerHTML = '<li class="empty-message">暂无文件</li>';
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

    listEl.innerHTML = files.map(f => {
        const isDir = Boolean(f.is_dir);
        const lastDotIndex = isDir ? -1 : f.file_name.lastIndexOf('.');
        const nameWithoutExt = lastDotIndex > -1 ? f.file_name.substring(0, lastDotIndex) : f.file_name;
        const extension = lastDotIndex > -1 ? f.file_name.substring(lastDotIndex) : '';
        const safeFileName = escapeHTML(f.file_name);
        const safeNameWithoutExt = escapeHTML(nameWithoutExt);
        const safeExtension = escapeHTML(extension);
        const kindText = isDir ? '文件夹' : '文件';
        const downloadName = isDir ? `${safeFileName}.zip` : safeFileName;
        
        return `
        <li class="file-item${isDir ? ' folder-item' : ''}">
            <div class="file-info">
                <div class="file-name-row">
                    <div class="file-name" id="file-name-${f.id}">${safeFileName}</div>
                    <div class="rename-input-container" id="rename-container-${f.id}">
                        <input type="text" id="rename-input-${f.id}" value="${safeNameWithoutExt}">
                        <span class="file-extension">${safeExtension}</span>
                    </div>
                    <div class="edit-buttons">
                        <a href="javascript:void(0);" class="edit-btn" id="edit-btn-${f.id}" onclick="startRename('${f.id}')">重命名</a>
                        <a href="javascript:void(0);" class="save-btn" id="save-btn-${f.id}" onclick="saveRename('${f.id}')">保存</a>
                        <a href="javascript:void(0);" class="cancel-btn" id="cancel-btn-${f.id}" onclick="cancelRename('${f.id}')">取消</a>
                    </div>
                </div>
                <div class="file-meta">
                    <span class="file-kind">${kindText}</span>
                    <span class="file-size">${formatFileSize(f.file_size)}</span>
                    <span class="file-time">更新时间：${new Date(f.update_time).toLocaleString()}</span>
                </div>
            </div>
            <div class="file-actions">
                <a href="/api/download/${f.id}" class="download-btn" download="${downloadName}">下载</a>
                <a href="javascript:void(0);" class="delete-btn" onclick="deleteFile('${f.id}', ${isDir})">删除</a>
            </div>
        </li>
    `;
    }).join('');
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

// 显示所选文件
function showSelectedFiles() {
    const selectedFilesContainer = document.getElementById('selected-files');
    
    if (selectedFilesList.length === 0) {
        selectedFilesContainer.innerHTML = '<div class="selected-files-empty">未选择文件</div>';
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
    
    let html = '<h3>所选文件 (' + selectedFilesList.length + ')</h3><ul>';
    for (let i = 0; i < selectedFilesList.length; i++) {
        const file = selectedFilesList[i];
        const uploadPath = getUploadPath(file);
        html += `
            <li class="selected-file-item">
                <span class="selected-file-name">${escapeHTML(uploadPath)}</span>
                <span class="selected-file-size">${formatFileSize(file.size)}</span>
                <button type="button" class="remove-file-btn" onclick="removeFile(${i})">移除</button>
            </li>
        `;
    }
    html += '</ul>';
    selectedFilesContainer.innerHTML = html;
}

// 移除单个文件
function removeFile(index) {
    selectedFilesList.splice(index, 1);
    showSelectedFiles();
}

// 初始化文件选择事件
function initFileInput() {
    const picker = document.getElementById('file-picker');
    const trigger = document.getElementById('file-picker-trigger');
    const selectFilesBtn = document.getElementById('select-files-btn');
    const selectFolderBtn = document.getElementById('select-folder-btn');
    const fileInput = document.getElementById('file-input');
    const folderInput = document.getElementById('folder-input');

    const closePicker = () => {
        picker.classList.remove('is-open');
        trigger.setAttribute('aria-expanded', 'false');
    };

    const togglePicker = () => {
        const isOpen = picker.classList.toggle('is-open');
        trigger.setAttribute('aria-expanded', String(isOpen));
    };

    const addFilesFromInput = (input) => {
        const files = input.files;
        for (let i = 0; i < files.length; i++) {
            const uploadPath = getUploadPath(files[i]);
            // 检查文件是否已经存在
            const isDuplicate = selectedFilesList.some(existingFile => 
                getUploadPath(existingFile) === uploadPath && existingFile.size === files[i].size
            );
            if (!isDuplicate) {
                selectedFilesList.push(files[i]);
            }
        }
        // 重置文件输入，以便可以再次选择相同的文件
        input.value = '';
        showSelectedFiles();
    };

    trigger.addEventListener('click', togglePicker);
    selectFilesBtn.addEventListener('click', () => {
        closePicker();
        fileInput.click();
    });
    selectFolderBtn.addEventListener('click', () => {
        closePicker();
        folderInput.click();
    });
    document.addEventListener('click', (event) => {
        if (!picker.contains(event.target)) {
            closePicker();
        }
    });
    document.addEventListener('keydown', (event) => {
        if (event.key === 'Escape') {
            closePicker();
        }
    });
    fileInput.addEventListener('change', () => addFilesFromInput(fileInput));
    folderInput.addEventListener('change', () => addFilesFromInput(folderInput));
}

// 初始化清空按钮
function initClearButton() {
    document.getElementById('clear-btn').addEventListener('click', () => {
        document.getElementById('file-input').value = '';
        document.getElementById('folder-input').value = '';
        document.getElementById('upload-status').textContent = '';
        selectedFilesList = [];
        showSelectedFiles();
    });
}

// 初始化
loadFiles();
initSearch();
initClearButton();
initFileInput();
// 初始化所选文件列表
showSelectedFiles();
