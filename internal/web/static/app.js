(function () {
  try {
    var t = localStorage.getItem('rt-theme');
    if (t) document.documentElement.dataset.theme = t;
  } catch (e) {}
  var btn = document.getElementById('theme');
  if (btn) {
    btn.addEventListener('click', function () {
      var r = document.documentElement;
      var dark = r.dataset.theme
        ? r.dataset.theme === 'dark'
        : window.matchMedia('(prefers-color-scheme: dark)').matches;
      r.dataset.theme = dark ? 'light' : 'dark';
      try { localStorage.setItem('rt-theme', r.dataset.theme); } catch (e) {}
    });
  }
})();
