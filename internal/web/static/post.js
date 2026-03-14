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
