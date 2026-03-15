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

// Update random/filter buttons on input
document.getElementById('posts-search').addEventListener('input', function() {
  updateRandomButton();
  updateTagFilterButtons();
});

// Extract last comma-separated term for autocomplete query
document.getElementById('posts-search').addEventListener('htmx:configRequest', function(e) {
  var value = this.value;
  var lastComma = value.lastIndexOf(',');
  var lastTerm = (lastComma >= 0 ? value.substring(lastComma + 1) : value).trim();
  e.detail.parameters.q = lastTerm;
});

// Handle autocomplete selection: replace last term
document.addEventListener('ac-select', function(e) {
  var search = document.getElementById('posts-search');
  if (e.target !== search) return;
  var v = search.value;
  var lastComma = v.lastIndexOf(',');
  search.value = (lastComma >= 0 ? v.substring(0, lastComma + 1) + ' ' : '') + e.detail.value;
});
