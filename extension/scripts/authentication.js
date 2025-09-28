let tokenPopup = document.querySelector('.token-popup')
tokenPopup.querySelector('.token-popup-done-btn').addEventListener('click', () => {
	let token = tokenPopup.querySelector('input').value
	if (token.trim() == '') {
		alert('Token is empty.')
		return
	}

	let saveTokenPort = chrome.runtime.connect({name: 'save-token'})
	saveTokenPort.postMessage({token: token})
	tokenPopup.close()
})

let isAuthenticatedPort = chrome.runtime.connect({name: 'is-authenticated'})
isAuthenticatedPort.onMessage.addListener((obj) => {
	if (!obj.isAuthenticated) {
		tokenPopup.showModal()
	}
})

