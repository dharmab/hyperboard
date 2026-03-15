(function() {
  var STORAGE_KEY = 'upload-queue';
  var form = document.getElementById('upload-form');
  var fileInput = document.getElementById('file-input');
  var previewZone = document.getElementById('preview-zone');
  var pasteZone = document.getElementById('paste-zone');

  var fileMap = new Map();
  var pendingHashes = new Set();
  var uploading = false;

  function hashFile(file) {
    return file.arrayBuffer().then(function(buf) {
      return crypto.subtle.digest('SHA-256', buf);
    }).then(function(hash) {
      var arr = new Uint8Array(hash);
      var hex = '';
      for (var i = 0; i < arr.length; i++) {
        hex += arr[i].toString(16).padStart(2, '0');
      }
      return hex;
    });
  }

  function fileToDataURL(file) {
    return new Promise(function(resolve) {
      var reader = new FileReader();
      reader.onload = function() { resolve(reader.result); };
      reader.readAsDataURL(file);
    });
  }

  function dataURLToFile(dataURL, name, type) {
    var parts = dataURL.split(',');
    var binary = atob(parts[1]);
    var arr = new Uint8Array(binary.length);
    for (var i = 0; i < binary.length; i++) {
      arr[i] = binary.charCodeAt(i);
    }
    return new File([arr], name, { type: type });
  }

  function saveQueue() {
    var items = [];
    fileMap.forEach(function(entry, key) {
      items.push({ key: key, name: entry.file.name, type: entry.file.type, dataURL: entry.dataURL });
    });
    try {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(items));
    } catch(e) {
      // sessionStorage full or unavailable — silently ignore
    }
  }

  function clearQueue() {
    try { sessionStorage.removeItem(STORAGE_KEY); } catch(e) {}
  }

  function addFile(file) {
    if (uploading) return;
    var placeholderKey = Symbol();
    pendingHashes.add(placeholderKey);

    Promise.all([hashFile(file), fileToDataURL(file)]).then(function(results) {
      var key = results[0];
      var dataURL = results[1];
      pendingHashes.delete(placeholderKey);
      if (fileMap.has(key)) return;
      renderPreview(file, key, dataURL);
      saveQueue();
    });
  }

  function renderPreview(file, key, dataURL) {
    var wrapper = document.createElement('div');
    wrapper.className = 'file-preview';

    var thumbDiv = document.createElement('div');
    thumbDiv.className = 'file-preview-thumb';
    if (file.type.startsWith('image/')) {
      var img = document.createElement('img');
      img.src = URL.createObjectURL(file);
      thumbDiv.appendChild(img);
    } else {
      var nameEl = document.createElement('div');
      nameEl.className = 'file-name';
      nameEl.textContent = file.name;
      thumbDiv.appendChild(nameEl);
    }
    wrapper.appendChild(thumbDiv);

    var infoDiv = document.createElement('div');
    infoDiv.className = 'file-preview-info';
    var fnameEl = document.createElement('div');
    fnameEl.textContent = file.name || 'pasted-image.png';
    infoDiv.appendChild(fnameEl);
    wrapper.appendChild(infoDiv);

    var btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'remove-btn';
    btn.textContent = '\u00d7';
    btn.addEventListener('click', function() {
      fileMap.delete(key);
      wrapper.remove();
      saveQueue();
    });
    wrapper.appendChild(btn);

    previewZone.appendChild(wrapper);
    fileMap.set(key, { file: file, element: wrapper, dataURL: dataURL });
  }

  // Restore queued files from sessionStorage
  (function restoreQueue() {
    var raw;
    try { raw = sessionStorage.getItem(STORAGE_KEY); } catch(e) { return; }
    if (!raw) return;
    var items;
    try { items = JSON.parse(raw); } catch(e) { return; }
    items.forEach(function(item) {
      if (fileMap.has(item.key)) return;
      var file = dataURLToFile(item.dataURL, item.name, item.type);
      renderPreview(file, item.key, item.dataURL);
    });
  })();

  fileInput.addEventListener('change', function() {
    for (var i = 0; i < fileInput.files.length; i++) {
      addFile(fileInput.files[i]);
    }
    fileInput.value = '';
  });

  var pasteZonePlaceholder = pasteZone.innerHTML;

  function handlePaste(e) {
    if (uploading) return;
    var items = e.clipboardData && e.clipboardData.items;
    if (!items) return;
    var handled = false;
    for (var i = 0; i < items.length; i++) {
      if (items[i].type.startsWith('image/')) {
        e.preventDefault();
        var file = items[i].getAsFile();
        if (file) addFile(file);
        handled = true;
      }
    }
    if (handled) {
      // Clear contenteditable div to prevent pasted image from rendering inline, then restore placeholder
      setTimeout(function() { pasteZone.innerHTML = pasteZonePlaceholder; }, 0);
    }
  }

  pasteZone.addEventListener('paste', handlePaste);
  document.addEventListener('paste', handlePaste);

  function mediaUrl(rawUrl) {
    try {
      var u = new URL(rawUrl);
      return '/media' + u.pathname;
    } catch(e) {
      return rawUrl;
    }
  }

  function uploadFile(entry, forceUpload) {
    var wrapper = entry.element;
    var infoDiv = wrapper.querySelector('.file-preview-info');

    // Remove previous status/similar content on re-upload
    var oldProgress = wrapper.querySelector('.progress-bar');
    if (oldProgress) oldProgress.remove();
    var oldStatus = wrapper.querySelector('.upload-status');
    if (oldStatus) oldStatus.remove();
    var oldSimilar = wrapper.querySelector('.similar-posts');
    if (oldSimilar) oldSimilar.remove();
    wrapper.classList.remove('upload-success', 'upload-error', 'upload-similar');

    // Remove the remove button during upload
    var removeBtn = wrapper.querySelector('.remove-btn');
    if (removeBtn) removeBtn.remove();

    // Add progress bar
    var progressBar = document.createElement('div');
    progressBar.className = 'progress-bar';
    var progressFill = document.createElement('div');
    progressFill.className = 'progress-fill';
    progressBar.appendChild(progressFill);
    infoDiv.appendChild(progressBar);

    // Add status text
    var statusEl = document.createElement('div');
    statusEl.className = 'upload-status';
    statusEl.textContent = 'Uploading\u2026';
    infoDiv.appendChild(statusEl);

    return new Promise(function(resolve) {
      var xhr = new XMLHttpRequest();
      var formData = new FormData();
      formData.append('files', entry.file, entry.file.name || 'pasted-image.png');
      if (forceUpload) {
        formData.append('force', 'true');
      }

      xhr.upload.addEventListener('progress', function(e) {
        if (e.lengthComputable) {
          var pct = Math.round((e.loaded / e.total) * 100);
          progressFill.style.width = pct + '%';
        }
      });

      xhr.addEventListener('load', function() {
        progressFill.style.width = '100%';
        var resp;
        try { resp = JSON.parse(xhr.responseText); } catch(e) { resp = {}; }

        if (xhr.status === 201 || (xhr.status >= 200 && xhr.status < 300 && resp.id)) {
          wrapper.classList.add('upload-success');
          var link = document.createElement('a');
          link.href = '/posts/' + resp.id;
          link.textContent = 'View post';
          statusEl.textContent = '';
          statusEl.appendChild(link);
          resolve();
        } else if (xhr.status === 409 && resp.similar && resp.similar.length > 0) {
          // Similar posts found — show them with force/discard options
          wrapper.classList.add('upload-similar');
          statusEl.textContent = 'Similar posts found:';

          var similarDiv = document.createElement('div');
          similarDiv.className = 'similar-posts';

          var grid = document.createElement('div');
          grid.className = 'similar-posts-grid';
          resp.similar.forEach(function(post) {
            var a = document.createElement('a');
            a.href = '/posts/' + post.id;
            a.target = '_blank';
            var thumb = document.createElement('img');
            thumb.className = 'similar-thumb';
            thumb.src = mediaUrl(post.thumbnailUrl);
            thumb.alt = 'Similar post';
            a.appendChild(thumb);
            grid.appendChild(a);
          });
          similarDiv.appendChild(grid);

          var actions = document.createElement('div');
          actions.className = 'similar-actions';
          var forceBtn = document.createElement('button');
          forceBtn.type = 'button';
          forceBtn.className = 'btn btn-primary';
          forceBtn.textContent = 'Upload anyway';
          forceBtn.addEventListener('click', function() {
            uploadFile(entry, true).then(resolve);
          });
          actions.appendChild(forceBtn);

          var discardBtn = document.createElement('button');
          discardBtn.type = 'button';
          discardBtn.className = 'btn btn-danger';
          discardBtn.textContent = 'Discard';
          discardBtn.addEventListener('click', function() {
            wrapper.classList.remove('upload-similar');
            wrapper.classList.add('upload-error');
            statusEl.textContent = 'Discarded';
            similarDiv.remove();
            resolve();
          });
          actions.appendChild(discardBtn);

          similarDiv.appendChild(actions);
          infoDiv.appendChild(similarDiv);
          // Do NOT resolve — wait for user action
        } else {
          wrapper.classList.add('upload-error');
          statusEl.textContent = resp.error || resp.message || 'Upload failed';
          resolve();
        }
      });

      xhr.addEventListener('error', function() {
        progressFill.style.width = '100%';
        wrapper.classList.add('upload-error');
        statusEl.textContent = 'Upload failed';
        resolve();
      });

      xhr.open('POST', '/upload');
      xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
      xhr.send(formData);
    });
  }

  form.addEventListener('submit', function(e) {
    e.preventDefault();
    if (fileMap.size === 0) return;

    uploading = true;
    fileInput.disabled = true;
    form.querySelector('button[type="submit"]').disabled = true;
    pasteZone.contentEditable = 'false';
    pasteZone.classList.add('disabled');

    var entries = [];
    fileMap.forEach(function(entry) { entries.push(entry); });

    // Upload files serially
    var chain = Promise.resolve();
    entries.forEach(function(entry) {
      chain = chain.then(function() { return uploadFile(entry, false); });
    });
    chain.then(function() {
      fileMap.clear();
      clearQueue();
    });
  });
})();
