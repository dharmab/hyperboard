(function() {
  var tables = document.querySelectorAll('table.data-table');
  tables.forEach(function(table) {
    var headers = table.querySelectorAll('th.sortable');
    headers.forEach(function(th) {
      th.style.cursor = 'pointer';
      th.addEventListener('click', function() {
        var col = th.cellIndex;
        var type = th.dataset.type;
        var wasUnsorted = !th.classList.contains('sorted-asc') && !th.classList.contains('sorted-desc');
        var asc;
        if (wasUnsorted) {
          asc = type !== 'number';
        } else {
          asc = !th.classList.contains('sorted-asc');
        }
        headers.forEach(function(h) { h.classList.remove('sorted-asc', 'sorted-desc'); });
        th.classList.add(asc ? 'sorted-asc' : 'sorted-desc');
        var tbody = table.querySelector('tbody');
        var rows = Array.from(tbody.querySelectorAll('tr'));
        rows.sort(function(a, b) {
          var av = a.cells[col].dataset.value;
          var bv = b.cells[col].dataset.value;
          var cmp;
          if (type === 'number') {
            cmp = parseInt(av, 10) - parseInt(bv, 10);
          } else {
            cmp = av.localeCompare(bv, undefined, {sensitivity: 'base'});
          }
          return asc ? cmp : -cmp;
        });
        rows.forEach(function(r) { tbody.appendChild(r); });
      });
    });
  });
})();
