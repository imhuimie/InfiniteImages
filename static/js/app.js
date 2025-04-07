// 全局变量
let currentPage = 1;
let totalPages = 1;
let imagesPerPage = 20;
let currentTag = '';
let apiKey = localStorage.getItem('apiKey') || '';

// DOM 元素
const dropArea = document.getElementById('drop-area');
const fileInput = document.getElementById('file-input');
const uploadForm = document.getElementById('upload-form');
const apiKeyInput = document.getElementById('api-key');
const uploadButton = document.getElementById('upload-button');
const uploadProgress = document.getElementById('upload-progress');
const progressFill = document.querySelector('.progress-fill');
const progressPercent = document.getElementById('progress-percent');
const uploadResult = document.getElementById('upload-result');
const resultContent = document.getElementById('result-content');
const gallery = document.getElementById('gallery');
const searchInput = document.getElementById('search-input');
const searchButton = document.getElementById('search-button');
const gridViewButton = document.getElementById('grid-view');
const listViewButton = document.getElementById('list-view');
const prevPageButton = document.getElementById('prev-page');
const nextPageButton = document.getElementById('next-page');
const pageInfo = document.getElementById('page-info');
const imageModal = document.getElementById('image-modal');
const modalImage = document.getElementById('modal-image');
const modalFilename = document.getElementById('modal-filename');
const modalDimensions = document.getElementById('modal-dimensions');
const modalSize = document.getElementById('modal-size');
const modalFormat = document.getElementById('modal-format');
const modalDate = document.getElementById('modal-date');
const modalTags = document.getElementById('modal-tags');
const tagInput = document.getElementById('tag-input');
const addTagButton = document.getElementById('add-tag-button');
const copyUrlButton = document.getElementById('copy-url-button');
const deleteImageButton = document.getElementById('delete-image-button');
const closeButton = document.querySelector('.close-button');

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    // 设置 API 密钥
    if (apiKey) {
        apiKeyInput.value = apiKey;
    }

    // 加载图片库
    loadGallery();

    // 设置事件监听器
    setupEventListeners();
});

// 设置事件监听器
function setupEventListeners() {
    // 拖放上传
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, preventDefaults, false);
    });

    ['dragenter', 'dragover'].forEach(eventName => {
        dropArea.addEventListener(eventName, highlight, false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, unhighlight, false);
    });

    dropArea.addEventListener('drop', handleDrop, false);

    // 表单提交
    uploadForm.addEventListener('submit', handleSubmit);

    // API 密钥保存
    apiKeyInput.addEventListener('change', () => {
        apiKey = apiKeyInput.value;
        localStorage.setItem('apiKey', apiKey);
    });

    // 视图切换
    gridViewButton.addEventListener('click', () => {
        gallery.className = 'grid-view';
        gridViewButton.classList.add('active');
        listViewButton.classList.remove('active');
        localStorage.setItem('viewMode', 'grid');
    });

    listViewButton.addEventListener('click', () => {
        gallery.className = 'list-view';
        listViewButton.classList.add('active');
        gridViewButton.classList.remove('active');
        localStorage.setItem('viewMode', 'list');
    });

    // 搜索
    searchButton.addEventListener('click', () => {
        currentTag = searchInput.value.trim();
        currentPage = 1;
        loadGallery();
    });

    searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            currentTag = searchInput.value.trim();
            currentPage = 1;
            loadGallery();
        }
    });

    // 分页
    prevPageButton.addEventListener('click', () => {
        if (currentPage > 1) {
            currentPage--;
            loadGallery();
        }
    });

    nextPageButton.addEventListener('click', () => {
        if (currentPage < totalPages) {
            currentPage++;
            loadGallery();
        }
    });

    // 模态框
    closeButton.addEventListener('click', () => {
        imageModal.classList.add('hidden');
    });

    window.addEventListener('click', (e) => {
        if (e.target === imageModal) {
            imageModal.classList.add('hidden');
        }
    });

    // 标签操作
    addTagButton.addEventListener('click', addTag);
    tagInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            addTag();
        }
    });

    // 复制 URL
    copyUrlButton.addEventListener('click', copyImageUrl);

    // 删除图片
    deleteImageButton.addEventListener('click', deleteImage);

    // 恢复视图模式
    const savedViewMode = localStorage.getItem('viewMode');
    if (savedViewMode === 'list') {
        gallery.className = 'list-view';
        listViewButton.classList.add('active');
        gridViewButton.classList.remove('active');
    } else {
        gallery.className = 'grid-view';
        gridViewButton.classList.add('active');
        listViewButton.classList.remove('active');
    }
}

// 阻止默认行为
function preventDefaults(e) {
    e.preventDefault();
    e.stopPropagation();
}

// 高亮拖放区域
function highlight() {
    dropArea.classList.add('highlight');
}

// 取消高亮拖放区域
function unhighlight() {
    dropArea.classList.remove('highlight');
}

// 处理拖放
function handleDrop(e) {
    const dt = e.dataTransfer;
    const files = dt.files;
    fileInput.files = files;
}

// 处理表单提交
function handleSubmit(e) {
    e.preventDefault();
    
    const files = fileInput.files;
    if (files.length === 0) {
        alert('请选择至少一个图片文件');
        return;
    }

    // 检查文件类型
    for (let i = 0; i < files.length; i++) {
        if (!files[i].type.startsWith('image/')) {
            alert('请只上传图片文件');
            return;
        }
    }

    // 保存 API 密钥
    apiKey = apiKeyInput.value;
    localStorage.setItem('apiKey', apiKey);

    // 显示进度条
    uploadProgress.classList.remove('hidden');
    uploadResult.classList.add('hidden');
    progressFill.style.width = '0%';
    progressPercent.textContent = '0%';

    // 创建表单数据
    const formData = new FormData();
    for (let i = 0; i < files.length; i++) {
        formData.append('images', files[i]);
    }

    // 发送请求
    const xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/upload');
    
    // 设置 API 密钥
    if (apiKey) {
        xhr.setRequestHeader('X-API-Key', apiKey);
    }

    // 进度事件
    xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
            const percent = Math.round((e.loaded / e.total) * 100);
            progressFill.style.width = percent + '%';
            progressPercent.textContent = percent + '%';
        }
    });

    // 完成事件
    xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
            const response = JSON.parse(xhr.responseText);
            showUploadResult(response);
            // 重新加载图片库
            loadGallery();
        } else {
            let errorMessage = '上传失败';
            try {
                const response = JSON.parse(xhr.responseText);
                errorMessage = response.message || errorMessage;
            } catch (e) {
                console.error('解析响应失败', e);
            }
            showUploadError(errorMessage);
        }
    });

    // 错误事件
    xhr.addEventListener('error', () => {
        showUploadError('网络错误，请稍后重试');
    });

    // 发送请求
    xhr.send(formData);
}

// 显示上传结果
function showUploadResult(response) {
    uploadResult.classList.remove('hidden');
    resultContent.innerHTML = '';

    if (response.success) {
        const urls = response.data;
        if (urls && urls.length > 0) {
            urls.forEach(url => {
                const item = document.createElement('div');
                item.className = 'result-item';
                
                const img = document.createElement('img');
                img.src = url;
                img.alt = '上传的图片';
                
                const info = document.createElement('div');
                info.className = 'result-info';
                
                const urlText = document.createElement('div');
                urlText.className = 'result-url';
                urlText.textContent = url;
                
                const copyBtn = document.createElement('button');
                copyBtn.className = 'copy-button';
                copyBtn.textContent = '复制链接';
                copyBtn.addEventListener('click', () => {
                    navigator.clipboard.writeText(window.location.origin + url)
                        .then(() => {
                            copyBtn.textContent = '已复制';
                            setTimeout(() => {
                                copyBtn.textContent = '复制链接';
                            }, 2000);
                        })
                        .catch(err => {
                            console.error('复制失败:', err);
                        });
                });
                
                info.appendChild(urlText);
                info.appendChild(copyBtn);
                
                item.appendChild(img);
                item.appendChild(info);
                
                resultContent.appendChild(item);
            });
        } else {
            resultContent.textContent = '上传成功，但未返回图片 URL';
        }
    } else {
        showUploadError(response.message || '上传失败');
    }
}

// 显示上传错误
function showUploadError(message) {
    uploadResult.classList.remove('hidden');
    resultContent.innerHTML = `<div class="error-message">${message}</div>`;
    
    if (Array.isArray(message.errors) && message.errors.length > 0) {
        const errorList = document.createElement('ul');
        message.errors.forEach(error => {
            const li = document.createElement('li');
            li.textContent = error;
            errorList.appendChild(li);
        });
        resultContent.appendChild(errorList);
    }
}

// 加载图片库
function loadGallery() {
    gallery.innerHTML = '<div class="loading">加载中...</div>';
    
    // 构建 URL
    let url = `/api/images?page=${currentPage}&limit=${imagesPerPage}`;
    if (currentTag) {
        url += `&tag=${encodeURIComponent(currentTag)}`;
    }
    
    // 发送请求
    fetch(url, {
        headers: apiKey ? { 'X-API-Key': apiKey } : {}
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('获取图片列表失败');
        }
        return response.json();
    })
    .then(data => {
        if (data.success) {
            displayGallery(data);
        } else {
            throw new Error(data.message || '获取图片列表失败');
        }
    })
    .catch(error => {
        gallery.innerHTML = `<div class="error-message">${error.message}</div>`;
    });
}

// 显示图片库
function displayGallery(data) {
    gallery.innerHTML = '';
    
    if (!data.data || data.data.length === 0) {
        gallery.innerHTML = '<div class="no-images">没有找到图片</div>';
        return;
    }
    
    // 更新分页信息
    totalPages = Math.ceil(data.total / data.limit);
    currentPage = data.page;
    imagesPerPage = data.limit;
    
    pageInfo.textContent = `第 ${currentPage} 页，共 ${totalPages} 页`;
    prevPageButton.disabled = currentPage <= 1;
    nextPageButton.disabled = currentPage >= totalPages;
    
    // 创建图片项
    data.data.forEach(image => {
        const item = document.createElement('div');
        item.className = 'gallery-item';
        item.dataset.id = image.id;
        
        const img = document.createElement('img');
        img.src = image.thumbnailUrl;
        img.alt = image.filename;
        img.loading = 'lazy';
        
        const info = document.createElement('div');
        info.className = 'item-info';
        
        if (gallery.className === 'grid-view') {
            // 网格视图
            const filename = document.createElement('div');
            filename.className = 'item-filename';
            filename.textContent = image.filename;
            
            const dimensions = document.createElement('div');
            dimensions.className = 'item-dimensions';
            dimensions.textContent = `${image.width} × ${image.height}`;
            
            info.appendChild(filename);
            info.appendChild(dimensions);
        } else {
            // 列表视图
            const details = document.createElement('div');
            details.className = 'item-details';
            
            const filename = document.createElement('div');
            filename.className = 'item-filename';
            filename.textContent = image.filename;
            
            const dimensions = document.createElement('div');
            dimensions.className = 'item-dimensions';
            dimensions.textContent = `${image.width} × ${image.height}`;
            
            details.appendChild(filename);
            details.appendChild(dimensions);
            
            const date = document.createElement('div');
            date.className = 'item-date';
            date.textContent = new Date(image.createdAt).toLocaleString();
            
            info.appendChild(details);
            info.appendChild(date);
        }
        
        item.appendChild(img);
        item.appendChild(info);
        
        // 点击事件
        item.addEventListener('click', () => {
            openImageModal(image);
        });
        
        gallery.appendChild(item);
    });
}

// 打开图片模态框
function openImageModal(image) {
    modalImage.src = image.url;
    modalFilename.textContent = image.filename;
    modalDimensions.textContent = `${image.width} × ${image.height}`;
    modalSize.textContent = formatFileSize(image.size);
    modalFormat.textContent = image.format.toUpperCase();
    modalDate.textContent = new Date(image.createdAt).toLocaleString();
    
    // 设置标签
    modalTags.innerHTML = '';
    if (image.tags && image.tags.length > 0) {
        image.tags.forEach(tag => {
            addTagElement(tag, image.id);
        });
    }
    
    // 设置当前图片 ID
    modalImage.dataset.id = image.id;
    modalImage.dataset.url = image.url;
    
    // 显示模态框
    imageModal.classList.remove('hidden');
}

// 添加标签元素
function addTagElement(tag, imageId) {
    const tagElement = document.createElement('div');
    tagElement.className = 'tag';
    tagElement.textContent = tag;
    
    const removeButton = document.createElement('span');
    removeButton.className = 'remove-tag';
    removeButton.innerHTML = '&times;';
    removeButton.addEventListener('click', (e) => {
        e.stopPropagation();
        removeTag(imageId, tag);
    });
    
    tagElement.appendChild(removeButton);
    modalTags.appendChild(tagElement);
}

// 添加标签
function addTag() {
    const tag = tagInput.value.trim();
    if (!tag) return;
    
    const imageId = modalImage.dataset.id;
    if (!imageId) return;
    
    // 获取当前标签
    const currentTags = Array.from(modalTags.querySelectorAll('.tag'))
        .map(el => el.textContent.replace('×', '').trim());
    
    // 检查标签是否已存在
    if (currentTags.includes(tag)) {
        tagInput.value = '';
        return;
    }
    
    // 添加标签
    fetch(`/api/images/${imageId}/tags`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Key': apiKey
        },
        body: JSON.stringify({
            tags: [...currentTags, tag]
        })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('更新标签失败');
        }
        return response.json();
    })
    .then(data => {
        if (data.success) {
            addTagElement(tag, imageId);
            tagInput.value = '';
            // 重新加载图片库
            loadGallery();
        } else {
            throw new Error(data.message || '更新标签失败');
        }
    })
    .catch(error => {
        alert(error.message);
    });
}

// 移除标签
function removeTag(imageId, tag) {
    // 获取当前标签
    const currentTags = Array.from(modalTags.querySelectorAll('.tag'))
        .map(el => el.textContent.replace('×', '').trim())
        .filter(t => t !== tag);
    
    // 更新标签
    fetch(`/api/images/${imageId}/tags`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Key': apiKey
        },
        body: JSON.stringify({
            tags: currentTags
        })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('更新标签失败');
        }
        return response.json();
    })
    .then(data => {
        if (data.success) {
            // 移除标签元素
            modalTags.querySelectorAll('.tag').forEach(el => {
                if (el.textContent.replace('×', '').trim() === tag) {
                    el.remove();
                }
            });
            // 重新加载图片库
            loadGallery();
        } else {
            throw new Error(data.message || '更新标签失败');
        }
    })
    .catch(error => {
        alert(error.message);
    });
}

// 复制图片 URL
function copyImageUrl() {
    const url = window.location.origin + modalImage.dataset.url;
    navigator.clipboard.writeText(url)
        .then(() => {
            copyUrlButton.textContent = '已复制';
            setTimeout(() => {
                copyUrlButton.textContent = '复制链接';
            }, 2000);
        })
        .catch(err => {
            console.error('复制失败:', err);
            alert('复制失败，请手动复制');
        });
}

// 删除图片
function deleteImage() {
    const imageId = modalImage.dataset.id;
    if (!imageId) return;
    
    if (!confirm('确定要删除这张图片吗？此操作不可撤销。')) {
        return;
    }
    
    fetch(`/api/images/${imageId}`, {
        method: 'DELETE',
        headers: {
            'X-API-Key': apiKey
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('删除图片失败');
        }
        return response.json();
    })
    .then(data => {
        if (data.success) {
            // 关闭模态框
            imageModal.classList.add('hidden');
            // 重新加载图片库
            loadGallery();
        } else {
            throw new Error(data.message || '删除图片失败');
        }
    })
    .catch(error => {
        alert(error.message);
    });
}

// 格式化文件大小
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}