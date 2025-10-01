let tokenPopup = document.querySelector('.token-popup')
tokenPopup.querySelector('.token-popup-done-btn').addEventListener('click', () => {
	let warnEl = tokenPopup.querySelector('.warn')
	let token = tokenPopup.querySelector('input').value

	if (token.trim() == '') {
		showEl(warnEl)
		warnEl.innerText = 'Token is empty'
		return
	}

	saveAuthToken(token).then((err) => {
		if (err != null) {
			showEl(warnEl)
			warnEl.innerText = err
			return
		}
		tokenPopup.close()
	})
})

async function saveAuthToken(token) {
	let res = await message.request('authentication.save-token', {
		token: token,
	})
	return res.error == undefined ? null : res.error
}

async function isAuthenticated() {
	let res = await message.request('authentication.is-authenticated')
	return res
}

isAuthenticated().then((res) => {
	if (!res.isAuthenticated) {
		tokenPopup.showModal()
	}
})
