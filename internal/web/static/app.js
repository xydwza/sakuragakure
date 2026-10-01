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
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/sw.js');
  }

  // pohon struktur pengurus: drag-to-pan (desktop), swipe native (mobile)
  var org = document.querySelector('.org-wrap');
  if (org) {
    var down = false, startX = 0, startScroll = 0;
    org.addEventListener('mousedown', function (e) {
      down = true; startX = e.pageX; startScroll = org.scrollLeft; org.classList.add('dragging');
    });
    window.addEventListener('mouseup', function () { down = false; org.classList.remove('dragging'); });
    org.addEventListener('mousemove', function (e) {
      if (!down) return;
      e.preventDefault();
      org.scrollLeft = startScroll - (e.pageX - startX);
    });
  }
})();
