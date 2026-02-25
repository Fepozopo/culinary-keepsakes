// Site UI helpers: theme persistence and UI bindings
(function () {
  const doc = document.documentElement;
  const saved = localStorage.getItem('theme');
  if (saved) doc.setAttribute('data-theme', saved);
  else if (!doc.getAttribute('data-theme')) doc.setAttribute('data-theme', 'dark');

  function initThemeToggle() {
    const btn = document.getElementById('theme-toggle');
    if (!btn) return;
    btn.addEventListener('click', () => {
      const next = doc.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
      doc.setAttribute('data-theme', next);
      try { localStorage.setItem('theme', next); } catch (e) { /* ignore */ }
    });
  }

  // Wait for DOM to be ready enough to find the toggle
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initThemeToggle);
  } else {
    initThemeToggle();
  }
})();
