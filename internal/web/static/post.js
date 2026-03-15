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
