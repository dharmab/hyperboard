// Show error banner with message, auto-dismiss after 8 seconds
function showError(msg) {
  var banner = document.getElementById('error-banner');
  banner.textContent = msg;
  banner.style.display = '';
  clearTimeout(banner._timeout);
  banner._timeout = setTimeout(function() { banner.style.display = 'none'; }, 8000);
}

// Handle HTMX HTTP error responses (4xx, 5xx)
document.body.addEventListener('htmx:responseError', function(e) {
  var xhr = e.detail.xhr;
  var msg = xhr.responseText || ('Request failed: HTTP ' + xhr.status);
  showError(msg.trim());
});

// Handle HTMX network/send errors
document.body.addEventListener('htmx:sendError', function(e) {
  showError('Network error: could not reach server');
});
