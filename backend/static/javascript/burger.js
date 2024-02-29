// Script
const container = document.querySelector('.navigation')
const primary = container.querySelector('.nav__list')
const primaryItems = container.querySelectorAll('.nav__list > li:not(.nav__item__more)')
container.classList.add('--jsfied')

// insert "more" button and duplicate the list
primary.insertAdjacentHTML('beforeend', `
  <li class="nav__item__more">
    <button type="button" aria-haspopup="true" aria-expanded="false">
      <svg id="fi_3018442" height="512" viewBox="0 0 24 24" width="512" xmlns="http://www.w3.org/2000/svg"><circle cx="12" cy="12" r="2"></circle><circle cx="4" cy="12" r="2"></circle><circle cx="20" cy="12" r="2"></circle></svg>
    </button>
    <ul class="nav__list__more">
      ${primary.innerHTML}
    </ul>
  </li>
`)

const secondary = container.querySelector('.nav__list__more')
const secondaryItems = secondary.querySelectorAll('li')
const allItems = container.querySelectorAll('li')
const moreLi = primary.querySelector('.nav__item__more')
const moreBtn = moreLi.querySelector('button')

moreBtn.addEventListener('click', (e) => {
  e.preventDefault()
  container.classList.toggle('nav__active')
  moreBtn.setAttribute('aria-expanded',     container.classList.contains('nav__active'))
})

// adapt tabs
const doAdapt = () => {
  // reveal all ite  ms for the calculation
  allItems.forEach((item) => {
    item.classList.remove('nav__hidden')
  })

  // hide items that won't fit in the Primary
  let stopWidth = moreBtn.offsetWidth
  let hiddenItems = []
  const primaryWidth = primary.offsetWidth
  primaryItems.forEach((item, i) => {
    if(primaryWidth >= stopWidth + item.offsetWidth) {
      stopWidth += item.offsetWidth
    } else {
      item.classList.add('nav__hidden')
      hiddenItems.push(i)
    }
  })
  
  // toggle the visibility of More button and items in Secondary
  if(!hiddenItems.length) {
    moreLi.classList.add('nav__hidden')
    container.classList.remove('nav__active')
    moreBtn.setAttribute('aria-expanded', false)
  }
  else {  
    secondaryItems.forEach((item, i) => {
      if(!hiddenItems.includes(i)) {
        item.classList.add('nav__hidden')
      }
    })
  }
}

doAdapt() // adapt immediately on load
window.addEventListener('resize', doAdapt) // adapt on window resize

// hide Secondary on the outside click

document.addEventListener('click', (e) => {
  let el = e.target
  while(el) {
    if(el === secondary || el === moreBtn) {
      return;
    }
    el = el.parentNode
  }
  container.classList.remove('nav__active')
  moreBtn.setAttribute('aria-expanded', false)
})