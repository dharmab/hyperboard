function setMediaMode(mode) {
  var el = document.getElementById('post-media');
  var fitBtn = document.getElementById('media-mode-fit');
  var fillBtn = document.getElementById('media-mode-fill');
  if (mode === 'fill') {
    el.classList.add('post-media--fill');
    fillBtn.classList.add('media-mode-option--active');
    fitBtn.classList.remove('media-mode-option--active');
  } else {
    el.classList.remove('post-media--fill');
    fitBtn.classList.add('media-mode-option--active');
    fillBtn.classList.remove('media-mode-option--active');
  }
  localStorage.setItem('mediaMode', mode);
}

(function() {
  if (localStorage.getItem('mediaMode') === 'fill') {
    setMediaMode('fill');
  }
})();

document.getElementById('media-mode-fit').addEventListener('click', function() {
  setMediaMode('fit');
});
document.getElementById('media-mode-fill').addEventListener('click', function() {
  setMediaMode('fill');
});

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

// Tag input Enter key handling
(function() {
  var tagInput = document.getElementById('tag-input');
  if (!tagInput) return;

  tagInput.addEventListener('keydown', function(event) {
    if (event.key === 'Enter') {
      event.preventDefault();
      this.closest('.post-tags-add').querySelector('button').click();
    }
  });
})();
