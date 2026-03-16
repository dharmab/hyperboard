// Note editor (only present when post has no note yet)
(function() {
  var addNoteBtn = document.getElementById('btn-add-note');
  if (!addNoteBtn) return;

  addNoteBtn.addEventListener('click', function() {
    var container = document.getElementById('post-note');
    var tpl = document.getElementById('note-editor-tpl');
    container.innerHTML = tpl.innerHTML;
    htmx.process(container);
    container.querySelector('textarea').focus();
  });
})();

// Tag autocomplete selection: click the Add button
document.addEventListener('ac-select', function(e) {
  var tagInput = document.getElementById('tag-input');
  if (e.target === tagInput) {
    tagInput.closest('.post-tags-add').querySelector('button').click();
  }
});

// Submit tag on Enter when no autocomplete item is active
document.addEventListener('keydown', function(e) {
  if (e.key !== 'Enter') return;
  var tagInput = document.getElementById('tag-input');
  if (!tagInput || e.target !== tagInput) return;
  if (document.querySelector('.ac-active')) return;
  if (!tagInput.value.trim()) return;
  e.preventDefault();
  tagInput.closest('.post-tags-add').querySelector('button').click();
});

// Keyboard shortcuts
(function() {
  function isInputFocused() {
    var el = document.activeElement;
    if (!el) return false;
    var tag = el.tagName.toLowerCase();
    return tag === 'input' || tag === 'textarea' || tag === 'select' || el.isContentEditable;
  }

  function getPostId() {
    var el = document.querySelector('[data-post-id]');
    return el ? el.getAttribute('data-post-id') : null;
  }

  document.addEventListener('keydown', function(e) {
    if (isInputFocused()) return;

    var postId = getPostId();
    if (!postId) return;

    // Arrow key navigation
    if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
      e.preventDefault();
      var el = document.querySelector('[data-created-at]');
      if (!el) return;
      var createdAt = el.getAttribute('data-created-at');
      var search;
      if (e.key === 'ArrowLeft') {
        search = 'sort:created,order:asc,created_after:' + createdAt;
      } else {
        search = 'sort:created,created_before:' + createdAt;
      }
      fetch('/search.json?limit=1&search=' + encodeURIComponent(search))
        .then(function(resp) {
          if (!resp.ok) return;
          return resp.json();
        })
        .then(function(data) {
          if (data && data.items && data.items.length > 0) {
            window.location.href = '/posts/' + data.items[0].id;
          }
        });
      return;
    }

    // Quick tag toggle with 'f'
    if (e.key === 'f') {
      var tagsEl = document.querySelector('.post-tags');
      if (!tagsEl) return;
      var quickTag = tagsEl.getAttribute('data-quick-tag');
      if (!quickTag) return;

      e.preventDefault();

      // Check if quick tag is among explicit (non-cascade) badges
      var badges = tagsEl.querySelectorAll('.badge:not(.badge-cascade) .badge-link');
      var hasTag = false;
      for (var i = 0; i < badges.length; i++) {
        if (badges[i].textContent.trim() === quickTag) {
          hasTag = true;
          break;
        }
      }

      if (hasTag) {
        htmx.ajax('DELETE', '/posts/' + postId + '/tags/' + encodeURIComponent(quickTag), {
          target: '.post-tags',
          swap: 'outerHTML'
        });
      } else {
        htmx.ajax('POST', '/posts/' + postId + '/tags', {
          target: '.post-tags',
          swap: 'outerHTML',
          values: { q: quickTag }
        });
      }
    }
  });
})();
