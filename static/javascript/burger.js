window.addEventListener(
  "pagehide",
  (event) => {
    if (document.getElementById('menu-checkbox')) {
      if (document.getElementById('menu-checkbox').checked) {
        localStorage.setItem('menuopen', true);
      } else {
        localStorage.removeItem('menuopen');
      }
    }
  },
  false,
);

if (document.getElementById('menu-checkbox')) {
  if (localStorage.getItem('menuopen')) {
    document.getElementById('menu-checkbox').checked=true;
  }
}

