// let toggle = document.querySelector('.toggle-darkmode');
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
