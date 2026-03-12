function enterEditMode() {
  document.getElementById('note-title-view').style.display = 'none';
  document.getElementById('note-title').style.display = '';
  document.getElementById('note-title').value = document.getElementById('note-title-view').textContent.trim();
  document.getElementById('note-edit').style.display = '';
  document.getElementById('note-view').style.display = 'none';
  document.getElementById('note-edit-btn').style.display = 'none';
  document.getElementById('note-save-btn').style.display = '';
}
function exitEditMode() {
  document.getElementById('note-title-view').textContent = document.getElementById('note-title').value;
  document.getElementById('note-title-view').style.display = '';
  document.getElementById('note-title').style.display = 'none';
  document.getElementById('note-edit').style.display = 'none';
  document.getElementById('note-view').style.display = '';
  document.getElementById('note-edit-btn').style.display = '';
  document.getElementById('note-save-btn').style.display = 'none';
}

document.getElementById('note-edit-btn').addEventListener('click', enterEditMode);
document.getElementById('note-save-btn').addEventListener('click', exitEditMode);
