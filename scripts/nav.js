document.querySelector('nav').addEventListener('click', (e) => {
	let navItem = e.target.closest('.nav-item')
    if (navItem == null) {
        return
	}
	document.querySelectorAll('.nav-item').forEach((el) => {
		el.classList.remove('selected-nav-item')
    })
    navItem.classList.add('selected-nav-item')
})
