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



let toggle = document.querySelector('#jour-nuit');

function lesgvGoDark(toggle) {
  localStorage.removeItem('lightmode');
  localStorage.setItem('darkmode', true);
  toggle.innerText = 'Nuit';
  document.body.classList.add('darkmode');
}

function lesgvGoLight(toggle) {
  localStorage.removeItem('darkmode');
  localStorage.setItem('lightmode', true);
  toggle.innerText = 'Jour';
  document.body.classList.remove('darkmode');
}

function toggleDarkmode() {
  let toggle = document.querySelector('#jour-nuit');
  if (document.body.classList.contains('darkmode')) {
    lesgvGoLight(toggle);
  } else {
    lesgvGoDark(toggle);
  }
}

toggle.addEventListener('click', function(e) {
  if (document.body.classList.contains('darkmode')) {
    lesgvGoLight(toggle);
  } else {
    lesgvGoDark(toggle);
  }
});

// Turn the theme off if the 'darkmode' key exists in localStorage
if (localStorage.getItem('darkmode')) {
  lesgvGoDark(toggle);
}  else if (localStorage.getItem('lightmode')) {
  lesgvGoLight(toggle);
} else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
  lesgvGoDark(toggle);
}


window.addEventListener(
  "pagehide",
  (event) => {
    if (toggle) {
      if (document.body.classList.contains('darkmode')) {
        lesgvGoDark(toggle);
      } else {
        lesgvGoLight(toggle);
      }
    }
  },
  false,
);