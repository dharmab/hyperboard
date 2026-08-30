// Show error banner with message, auto-dismiss after 8 seconds
function showError(msg) {
  var banner = document.getElementById('error-banner');
  banner.textContent = msg;
  banner.style.display = '';
  clearTimeout(banner._timeout);
  banner._timeout = setTimeout(function() { banner.style.display = 'none'; }, 8000);
}

// Handle HTMX HTTP error responses (4xx, 5xx)
document.body.addEventListener('htmx:response:error', function(e) {
  var ctx = e.detail.ctx;
  ctx.swap = 'none';
  var msg = ctx.text || ('Request failed: HTTP ' + ctx.response.status);
  showError(msg.trim());
});

// Handle HTMX request and swap failures
document.body.addEventListener('htmx:error', function(e) {
  var error = e.detail.error;
  var msg = error && (error.message || String(error));
  showError(msg || 'An unexpected error occurred');
});
