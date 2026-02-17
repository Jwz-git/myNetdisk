const deleteFile = async (id) => {
    if (!confirm('确定要删除这个文件吗？')) return;
    fetch(`/api/delete/${id}`, { method: 'DELETE' })  // 建议使用 DELETE 方法
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                window.location.href = '/admin'; // 手动跳转
            } else {
                alert('删除失败');
            }
        });

    loadFiles(); // 刷新列表
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
        formData.append('files', selectedFilesList[i]);
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
function startRename(id, nameWithoutExt, extension) {
    // 隐藏文件名和重命名按钮
    document.getElementById(`file-name-${id}`).style.display = 'none';
    document.querySelector(`#save-btn-${id}`).style.display = 'none';
    document.querySelector(`#cancel-btn-${id}`).style.display = 'none';
    document.querySelector(`.edit-btn[onclick="startRename('${id}', '${nameWithoutExt}', '${extension}')"]`).style.display = 'none';
    
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
    const editBtn = document.querySelector(`.edit-btn[onclick^="startRename('${id}'"]`);
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
    const res = await fetch('/api/files');
    allFiles = await res.json();
    const listEl = document.getElementById('file-list');
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
        const lastDotIndex = f.file_name.lastIndexOf('.');
        const nameWithoutExt = lastDotIndex > -1 ? f.file_name.substring(0, lastDotIndex) : f.file_name;
        const extension = lastDotIndex > -1 ? f.file_name.substring(lastDotIndex) : '';
        
        return `
        <li class="file-item">
            <div class="file-info">
                <div style="display:flex; align-items: center;">
                    <div class="file-name" id="file-name-${f.id}">${f.file_name}</div>
                    <div class="rename-input-container" id="rename-container-${f.id}" style="display: none; align-items: center;">
                        <input type="text" id="rename-input-${f.id}" value="${nameWithoutExt}" style="padding: 4px 8px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; margin-right: 8px;">
                        <span class="file-extension" style="font-size: 14px; color: #6c757d;">${extension}</span>
                    </div>
                    <div class="edit-buttons">
                        <a href="javascript:void(0);" class="edit-btn" onclick="startRename('${f.id}', '${nameWithoutExt}', '${extension}')">重命名</a>
                        <a href="javascript:void(0);" class="save-btn" id="save-btn-${f.id}" style="display: none; margin-left: 8px;" onclick="saveRename('${f.id}')">保存</a>
                        <a href="javascript:void(0);" class="cancel-btn" id="cancel-btn-${f.id}" style="display: none; margin-left: 8px;" onclick="cancelRename('${f.id}')">取消</a>
                    </div>
                </div>
                <div class="file-meta">
                    <span class="file-size">${formatFileSize(f.file_size)}</span>
                    <span class="file-time">更新时间：${new Date(f.update_time).toLocaleString()}</span>
                </div>
            </div>
            <div class="file-actions">
                <a href="/api/download/${f.id}" class="download-btn" download="${f.file_name}">下载</a>
                <a href="javascript:void(0);" class="delete-btn" onclick="deleteFile('${f.id}')">删除</a>
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
        html += `
            <li class="selected-file-item">
                <span class="selected-file-name">${file.name}</span>
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
    const fileInput = document.getElementById('file-input');
    fileInput.addEventListener('change', () => {
        const files = fileInput.files;
        for (let i = 0; i < files.length; i++) {
            // 检查文件是否已经存在
            const isDuplicate = selectedFilesList.some(existingFile => 
                existingFile.name === files[i].name && existingFile.size === files[i].size
            );
            if (!isDuplicate) {
                selectedFilesList.push(files[i]);
            }
        }
        // 重置文件输入，以便可以再次选择相同的文件
        fileInput.value = '';
        showSelectedFiles();
    });
}

// 初始化清空按钮
function initClearButton() {
    document.getElementById('clear-btn').addEventListener('click', () => {
        document.getElementById('file-input').value = '';
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