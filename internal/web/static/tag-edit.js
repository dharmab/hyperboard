(function() {
  // Confirm dialog for convert-to-alias form
  var convertForm = document.getElementById('convert-to-alias-form');
  if (convertForm) {
    convertForm.addEventListener('submit', function(e) {
      if (!confirm(this.dataset.confirmMessage)) {
        e.preventDefault();
      }
    });
  }

  // HTMX config for tag suggestion autocomplete
  var convertInput = document.getElementById('convert-target-input');
  if (convertInput) {
    convertInput.addEventListener('htmx:configRequest', function(e) {
      e.detail.parameters.q = this.value;
    });
  }
})();
