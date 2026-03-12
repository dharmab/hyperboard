function toggleRandom() {
  var input = document.getElementById('posts-search');
  var terms = input.value.split(',').map(function(s) { return s.trim(); }).filter(function(s) { return s !== '' && !s.startsWith('sort:'); });
  if (!hasRandomTerm(input.value)) {
    terms.push('sort:random');
  }
  input.value = terms.join(', ');
  htmx.trigger(document.querySelector('.posts-controls'), 'submit');
}

function hasRandomTerm(value) {
  return value.split(',').map(function(s) { return s.trim(); }).indexOf('sort:random') >= 0;
}

function updateRandomButton() {
  var input = document.getElementById('posts-search');
  var btn = document.getElementById('btn-random');
  if (hasRandomTerm(input.value)) {
    btn.classList.add('btn-primary');
  } else {
    btn.classList.remove('btn-primary');
  }
}
updateRandomButton();

function cycleTagFilter(btn) {
  var tags = JSON.parse(btn.dataset.tags);
  var input = document.getElementById('posts-search');
  var terms = input.value.split(',').map(function(s) { return s.trim(); }).filter(function(s) { return s !== ''; });

  var activeIndex = -1;
  for (var i = 0; i < tags.length; i++) {
    var idx = terms.indexOf(tags[i]);
    if (idx >= 0) {
      activeIndex = i;
      terms.splice(idx, 1);
      break;
    }
  }

  if (activeIndex === -1) {
    terms.push(tags[0]);
  } else if (activeIndex < tags.length - 1) {
    terms.push(tags[activeIndex + 1]);
  }

  input.value = terms.join(', ');
  htmx.trigger(document.querySelector('.posts-controls'), 'submit');
  updateTagFilterButtons();
}

function updateTagFilterButtons() {
  var input = document.getElementById('posts-search');
  var terms = input.value.split(',').map(function(s) { return s.trim(); });
  var btns = document.querySelectorAll('.tag-filter-btn');
  for (var i = 0; i < btns.length; i++) {
    var btnTags = JSON.parse(btns[i].dataset.tags);
    var active = false;
    for (var j = 0; j < btnTags.length; j++) {
      if (terms.indexOf(btnTags[j]) >= 0) {
        active = true;
        break;
      }
    }
    if (active) {
      btns[i].classList.add('btn-primary');
    } else {
      btns[i].classList.remove('btn-primary');
    }
  }
}
updateTagFilterButtons();

document.getElementById('btn-random').addEventListener('click', toggleRandom);

document.querySelectorAll('.tag-filter-btn').forEach(function(btn) {
  btn.addEventListener('click', function() { cycleTagFilter(this); });
});

(function() {
  var searchInput = document.getElementById('posts-search');
  var datalist = document.getElementById('search-suggestions');
  var debounceTimer;

  searchInput.addEventListener('input', function() {
    updateRandomButton();
    updateTagFilterButtons();
    clearTimeout(debounceTimer);
    var value = this.value;
    var lastCommaIdx = value.lastIndexOf(',');
    var lastWord = (lastCommaIdx >= 0 ? value.substring(lastCommaIdx + 1) : value).trim();

    if (!lastWord) {
      datalist.innerHTML = '';
      return;
    }

    debounceTimer = setTimeout(function() {
      fetch('/tag-suggestions?q=' + encodeURIComponent(lastWord))
        .then(function(r) { return r.text(); })
        .then(function(html) {
          if (lastCommaIdx >= 0) {
            var prefix = value.substring(0, lastCommaIdx + 1) + ' ';
            var temp = document.createElement('div');
            temp.innerHTML = html;
            temp.querySelectorAll('option').forEach(function(opt) {
              opt.value = prefix + opt.value;
            });
            datalist.innerHTML = temp.innerHTML;
          } else {
            datalist.innerHTML = html;
          }
        });
    }, 200);
  });
})();
