import { createZipBlob } from './zip.js';

const escapeHTML = (value) => String(value).replace(/[&<>"']/g, (char) => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
}[char]));

const getUploadPath = (file) => file.webkitRelativePath || file.name;

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
let selectedFilesList = [];

function setUploadStatus(message, state = 'idle') {
    const status = document.getElementById('upload-status');
    status.textContent = message;
    status.dataset.state = state;
}

function notify(message, type = 'success') {
    const region = document.getElementById('toast-region');
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.innerHTML = `
        <span class="toast-dot" aria-hidden="true"></span>
        <span>${escapeHTML(message)}</span>
    `;
    region.appendChild(toast);

    window.setTimeout(() => {
        toast.classList.add('is-leaving');
        toast.addEventListener('animationend', () => toast.remove(), { once: true });
    }, 2800);
}

function updateFileStats() {
    const totalSize = allFiles.reduce((sum, file) => sum + (Number(file.file_size) || 0), 0);
    document.getElementById('file-total').textContent = String(allFiles.length);
    document.getElementById('storage-total').textContent = formatFileSize(totalSize);
}

function updateResultCount(count) {
    document.getElementById('result-count').textContent = String(count);
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
        listEl.innerHTML = '<li class="empty-message error-message">文件目录暂时无法加载，请稍后刷新</li>';
        updateFileStats();
        updateResultCount(0);
        listEl.setAttribute('aria-busy', 'false');
        return;
    }

    updateFileStats();
    renderFiles(allFiles, '目录还是空的，从左侧上传第一个文件吧');
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
        const fileName = String(file.file_name);
        const lastDotIndex = isDir ? -1 : fileName.lastIndexOf('.');
        const nameWithoutExt = lastDotIndex > 0 ? fileName.substring(0, lastDotIndex) : fileName;
        const extension = lastDotIndex > 0 ? fileName.substring(lastDotIndex) : '';
        const safeFileName = escapeHTML(fileName);
        const safeNameWithoutExt = escapeHTML(nameWithoutExt);
        const safeExtension = escapeHTML(extension);
        const safeID = escapeHTML(file.id);
        const category = getFileCategory(fileName, isDir);
        const kindText = getKindLabel(fileName, isDir);
        const downloadName = isDir ? `${safeFileName}.zip` : safeFileName;
        const downloadURL = `/api/download/${encodeURIComponent(file.id)}`;

        return `
            <li class="file-item${isDir ? ' folder-item' : ''}" data-file-id="${safeID}" data-is-dir="${isDir}">
                <span class="file-icon file-icon-${category}" aria-hidden="true">${getFileIcon(category)}</span>
                <div class="file-info">
                    <div class="file-name-row">
                        <div class="file-name" title="${safeFileName}">${safeFileName}</div>
                        <div class="rename-input-container">
                            <input type="text" value="${safeNameWithoutExt}" aria-label="新的文件名">
                            <span class="file-extension">${safeExtension}</span>
                        </div>
                        <div class="edit-buttons">
                            <button type="button" class="edit-btn" data-action="rename" aria-label="重命名 ${safeFileName}">重命名</button>
                            <button type="button" class="save-btn" data-action="save" aria-label="保存 ${safeFileName} 的新名称">保存</button>
                            <button type="button" class="cancel-btn" data-action="cancel" aria-label="取消重命名 ${safeFileName}">取消</button>
                        </div>
                    </div>
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
                    <button type="button" class="delete-btn" data-action="delete" aria-label="删除 ${safeFileName}">
                        <svg viewBox="0 0 18 18" aria-hidden="true"><path d="M3.5 5h11M7 5V3h4v2m-6 0 .8 10h6.4L13 5M7.5 8v4.5m3-4.5v4.5"></path></svg>
                        删除
                    </button>
                </div>
            </li>
        `;
    }).join('');
}

function startRename(button) {
    const row = button.closest('.file-item');
    row.classList.add('is-renaming');
    const input = row.querySelector('.rename-input-container input');
    input.focus();
    input.select();
}

function cancelRename(button) {
    button.closest('.file-item').classList.remove('is-renaming');
}

async function saveRename(button) {
    const row = button.closest('.file-item');
    const input = row.querySelector('.rename-input-container input');
    const extension = row.querySelector('.file-extension').textContent;
    const newNameWithoutExt = input.value.trim();

    if (!newNameWithoutExt) {
        notify('文件名不能为空', 'error');
        input.focus();
        return;
    }

    button.disabled = true;
    try {
        const response = await fetch(`/api/rename/${encodeURIComponent(row.dataset.fileId)}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ file_name: newNameWithoutExt + extension })
        });
        const result = await response.json();
        if (!result.success) throw new Error(result.error || '未知错误');

        notify('重命名已保存');
        await loadFiles();
    } catch (error) {
        notify(`重命名失败：${error.message}`, 'error');
        button.disabled = false;
    }
}

function confirmDelete(isDir) {
    const dialog = document.getElementById('confirm-dialog');
    const message = isDir
        ? '这个文件夹及其中的全部内容都会被永久删除。'
        : '这个文件会被永久删除，此操作无法撤销。';

    if (typeof dialog.showModal !== 'function') {
        return Promise.resolve(window.confirm(message));
    }

    document.getElementById('confirm-message').textContent = message;
    dialog.returnValue = '';
    return new Promise((resolve) => {
        dialog.addEventListener('close', () => resolve(dialog.returnValue === 'confirm'), { once: true });
        dialog.showModal();
    });
}

async function deleteFile(button) {
    const row = button.closest('.file-item');
    const confirmed = await confirmDelete(row.dataset.isDir === 'true');
    if (!confirmed) return;

    button.disabled = true;
    try {
        const response = await fetch(`/api/delete/${encodeURIComponent(row.dataset.fileId)}`, { method: 'DELETE' });
        const result = response.headers.get('Content-Type')?.includes('application/json')
            ? await response.json()
            : { success: false, error: await response.text() };
        if (!result.success) throw new Error(result.error || '未知错误');

        notify('文件已删除');
        await loadFiles();
    } catch (error) {
        notify(`删除失败：${error.message}`, 'error');
        button.disabled = false;
    }
}

function initFileActions() {
    document.getElementById('file-list').addEventListener('click', (event) => {
        const button = event.target.closest('button[data-action]');
        if (!button) return;

        const actions = {
            rename: startRename,
            cancel: cancelRename,
            save: saveRename,
            delete: deleteFile
        };
        actions[button.dataset.action]?.(button);
    });

    document.getElementById('file-list').addEventListener('keydown', (event) => {
        if (event.key !== 'Enter') return;
        const input = event.target.closest('.rename-input-container input');
        if (input) input.closest('.file-item').querySelector('.save-btn').click();
    });
}

function searchFiles() {
    const searchTerm = document.getElementById('search-input').value.toLowerCase().trim();
    if (!searchTerm) {
        renderFiles(allFiles, '目录还是空的，从左侧上传第一个文件吧');
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
    renderFiles(allFiles, '目录还是空的，从左侧上传第一个文件吧');
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

function showSelectedFiles() {
    const container = document.getElementById('selected-files');
    if (selectedFilesList.length === 0) {
        container.innerHTML = '<div class="selected-files-empty">尚未选择文件</div>';
        return;
    }

    const totalSize = selectedFilesList.reduce((sum, file) => sum + file.size, 0);
    const folderCount = new Set(
        selectedFilesList
            .map((file) => file.webkitRelativePath?.split('/')[0])
            .filter(Boolean)
    ).size;
    const items = selectedFilesList.map((file, index) => `
        <li class="selected-file-item">
            <span class="selected-file-name" title="${escapeHTML(getUploadPath(file))}">${escapeHTML(getUploadPath(file))}</span>
            <span class="selected-file-size">${formatFileSize(file.size)}</span>
            <button type="button" class="remove-file-btn" data-index="${index}" aria-label="移除 ${escapeHTML(file.name)}">移除</button>
        </li>
    `).join('');

    const folderNotice = folderCount > 0
        ? `<p class="folder-zip-notice">${folderCount} 个文件夹将在浏览器内压缩为 ZIP，再上传到服务器</p>`
        : '';
    container.innerHTML = `<h3>已选择 ${selectedFilesList.length} 个文件 · ${formatFileSize(totalSize)}</h3>${folderNotice}<ul>${items}</ul>`;
}

function addFiles(files) {
    Array.from(files).forEach((file) => {
        const uploadPath = getUploadPath(file);
        const isDuplicate = selectedFilesList.some((existingFile) =>
            getUploadPath(existingFile) === uploadPath && existingFile.size === file.size
        );
        if (!isDuplicate) selectedFilesList.push(file);
    });

    showSelectedFiles();
    if (selectedFilesList.length > 0) {
        setUploadStatus(`已准备 ${selectedFilesList.length} 个项目`, 'idle');
    }
}

function buildUploadPlan(files) {
    const plans = [];
    const folderPlans = new Map();

    Array.from(files).forEach((file) => {
        const relativePath = file.webkitRelativePath || '';
        const separatorIndex = relativePath.indexOf('/');
        if (separatorIndex <= 0) {
            plans.push({ type: 'file', file, path: file.name });
            return;
        }

        const folderName = relativePath.slice(0, separatorIndex);
        let plan = folderPlans.get(folderName);
        if (!plan) {
            plan = { type: 'folder', folderName, files: [] };
            folderPlans.set(folderName, plan);
            plans.push(plan);
        }
        plan.files.push(file);
    });

    return plans;
}

async function prepareUploadItems(files) {
    const plans = buildUploadPlan(files);
    const uploadItems = [];
    const uploadPaths = new Set();

    for (const plan of plans) {
        let item;
        if (plan.type === 'file') {
            item = { file: plan.file, path: plan.path };
        } else {
            setUploadStatus(`正在压缩 ${plan.folderName}（0/${plan.files.length}）…`, 'uploading');
            const zipBlob = await createZipBlob(plan.files, ({ completed, total }) => {
                setUploadStatus(`正在压缩 ${plan.folderName}（${completed}/${total}）…`, 'uploading');
            });
            const zipName = `${plan.folderName}.zip`;
            item = {
                file: new File([zipBlob], zipName, { type: 'application/zip', lastModified: Date.now() }),
                path: zipName
            };
        }

        if (uploadPaths.has(item.path)) {
            throw new Error(`上传队列中存在同名项目：${item.path}`);
        }
        uploadPaths.add(item.path);
        uploadItems.push(item);
    }

    return uploadItems;
}

function setUploadBusy(isBusy) {
    const form = document.getElementById('upload-form');
    form.querySelector('.upload-btn').disabled = isBusy;
    document.getElementById('clear-btn').disabled = isBusy;
    document.getElementById('file-picker-trigger').disabled = isBusy;
    document.getElementById('select-files-btn').disabled = isBusy;
    document.getElementById('select-folder-btn').disabled = isBusy;
    document.getElementById('drop-zone').inert = isBusy;
    document.getElementById('selected-files').inert = isBusy;

    if (isBusy) {
        document.getElementById('file-picker').classList.remove('is-open');
        document.getElementById('file-picker-trigger').setAttribute('aria-expanded', 'false');
    }
}

function initFileInput() {
    const picker = document.getElementById('file-picker');
    const trigger = document.getElementById('file-picker-trigger');
    const menu = document.getElementById('file-picker-menu');
    const menuButtons = Array.from(menu.querySelectorAll('button'));
    const selectFilesBtn = document.getElementById('select-files-btn');
    const selectFolderBtn = document.getElementById('select-folder-btn');
    const fileInput = document.getElementById('file-input');
    const folderInput = document.getElementById('folder-input');
    const dropZone = document.getElementById('drop-zone');

    const closePicker = () => {
        picker.classList.remove('is-open');
        trigger.setAttribute('aria-expanded', 'false');
    };

    const togglePicker = () => {
        const isOpen = picker.classList.toggle('is-open');
        trigger.setAttribute('aria-expanded', String(isOpen));
        if (isOpen) menuButtons[0].focus();
    };

    const addFilesFromInput = (input) => {
        addFiles(input.files);
        input.value = '';
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
        if (!picker.contains(event.target)) closePicker();
    });
    document.addEventListener('keydown', (event) => {
        if (event.key === 'Escape') {
            closePicker();
            dropZone.classList.remove('is-dragging');
        }
    });
    menu.addEventListener('keydown', (event) => {
        if (!['ArrowDown', 'ArrowUp'].includes(event.key)) return;
        event.preventDefault();
        const currentIndex = menuButtons.indexOf(document.activeElement);
        const direction = event.key === 'ArrowDown' ? 1 : -1;
        menuButtons[(currentIndex + direction + menuButtons.length) % menuButtons.length].focus();
    });
    fileInput.addEventListener('change', () => addFilesFromInput(fileInput));
    folderInput.addEventListener('change', () => addFilesFromInput(folderInput));

    ['dragenter', 'dragover'].forEach((eventName) => {
        dropZone.addEventListener(eventName, (event) => {
            event.preventDefault();
            dropZone.classList.add('is-dragging');
        });
    });
    ['dragleave', 'drop'].forEach((eventName) => {
        dropZone.addEventListener(eventName, (event) => {
            event.preventDefault();
            dropZone.classList.remove('is-dragging');
        });
    });
    dropZone.addEventListener('drop', (event) => addFiles(event.dataTransfer.files));

    document.getElementById('selected-files').addEventListener('click', (event) => {
        const button = event.target.closest('.remove-file-btn');
        if (!button) return;
        selectedFilesList.splice(Number(button.dataset.index), 1);
        showSelectedFiles();
        setUploadStatus(selectedFilesList.length ? `已准备 ${selectedFilesList.length} 个项目` : '等待选择文件');
    });
}

function initUpload() {
    const form = document.getElementById('upload-form');

    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        if (selectedFilesList.length === 0) {
            setUploadStatus('请先选择要上传的内容', 'error');
            return;
        }

        const selectedSnapshot = [...selectedFilesList];
        let uploadItems = [];
        setUploadBusy(true);
        try {
            uploadItems = await prepareUploadItems(selectedSnapshot);
            const formData = new FormData();
            uploadItems.forEach((item) => {
                formData.append('files', item.file);
                formData.append('paths', item.path);
            });

            setUploadStatus(`正在上传 ${uploadItems.length} 个文件或压缩包…`, 'uploading');
            const response = await fetch('/api/upload', { method: 'POST', body: formData });
            const result = await response.json();
            if (!response.ok || !result.success) throw new Error(result.error || '未知错误');

            selectedFilesList = [];
            showSelectedFiles();
            setUploadStatus('上传完成，文件目录已更新', 'success');
            notify('上传成功');
            await loadFiles();
        } catch (error) {
            setUploadStatus(`上传失败：${error.message}`, 'error');
            notify('上传没有完成，请重试', 'error');
        } finally {
            uploadItems = [];
            setUploadBusy(false);
        }
    });

    document.getElementById('clear-btn').addEventListener('click', () => {
        document.getElementById('file-input').value = '';
        document.getElementById('folder-input').value = '';
        selectedFilesList = [];
        showSelectedFiles();
        setUploadStatus('等待选择文件');
    });
}

document.getElementById('current-year').textContent = String(new Date().getFullYear());
loadFiles();
initSearch();
initFileActions();
initFileInput();
initUpload();
showSelectedFiles();
