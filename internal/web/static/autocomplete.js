// Autocomplete keyboard navigation and selection for .ac-dropdown widgets
(function() {
  function getItems(dropdown) {
    return dropdown.querySelectorAll('.ac-item');
  }

  function clearDropdown(dropdown) {
    dropdown.innerHTML = '';
  }

  function selectItem(input, value) {
    input.value = value;
    input.dispatchEvent(new CustomEvent('ac-select', { bubbles: true, detail: { value: value } }));
    var dropdown = input.parentElement.querySelector('.ac-dropdown');
    if (dropdown) clearDropdown(dropdown);
  }

  document.body.addEventListener('keydown', function(e) {
    var input = e.target;
    if (input.tagName !== 'INPUT') return;
    var dropdown = input.parentElement.querySelector('.ac-dropdown');
    if (!dropdown) return;
    var items = getItems(dropdown);
    if (items.length === 0) return;

    if (e.key === 'Tab') {
      e.preventDefault();
      var active = dropdown.querySelector('.ac-active');
      if (active) {
        active.classList.remove('ac-active');
        var next = active.nextElementSibling || items[0];
        next.classList.add('ac-active');
        next.scrollIntoView({ block: 'nearest' });
      } else {
        items[0].classList.add('ac-active');
        items[0].scrollIntoView({ block: 'nearest' });
      }
    } else if (e.key === 'Enter') {
      var active = dropdown.querySelector('.ac-active');
      if (active) {
        e.preventDefault();
        selectItem(input, active.dataset.value);
      }
    } else if (e.key === 'Escape') {
      clearDropdown(dropdown);
    }
  });

  document.body.addEventListener('click', function(e) {
    var item = e.target.closest('.ac-item');
    if (!item) return;
    var dropdown = item.closest('.ac-dropdown');
    if (!dropdown) return;
    var input = dropdown.parentElement.querySelector('input');
    if (input) selectItem(input, item.dataset.value);
  });

  document.body.addEventListener('focusout', function(e) {
    var input = e.target;
    if (input.tagName !== 'INPUT') return;
    var dropdown = input.parentElement.querySelector('.ac-dropdown');
    if (!dropdown) return;
    setTimeout(function() { clearDropdown(dropdown); }, 150);
  });
})();
