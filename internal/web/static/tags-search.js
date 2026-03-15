(function() {
  var input = document.getElementById('tags-search');
  var tbody = document.querySelector('#tags-table tbody');
  function filterRows() {
    var q = input.value.toLowerCase();
    var rows = tbody.querySelectorAll('tr');
    rows.forEach(function(row) {
      var name = row.cells[0].dataset.value.toLowerCase();
      row.style.display = name.includes(q) ? '' : 'none';
    });
  }
  input.addEventListener('input', filterRows);
  input.addEventListener('ac-select', function(e) {
    input.value = e.detail.value;
    filterRows();
  });
})();
