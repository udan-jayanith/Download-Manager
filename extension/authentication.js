let authenticateUser = {
	saveToken: async function (token) {
		let settings = await getSettings()
		if (settings.cookies == undefined) {
			settings.cookies = {}
		}
		settings.cookies.token = token
		await updateSettings(settings)
	},
	isAuthenticated: async function () {
		let settings = await getSettings()
		let cookie = settings.cookies.token
		if (!cookie || cookie == null || cookie == '') {
			return false
		}
		return true
	},
}

chrome.runtime.onConnect.onPort('save-token', (port) => {
	port.onMessage.addListener((obj) => {
		authenticateUser.saveToken(obj.token)
	})
})

chrome.runtime.onConnect.onPort('is-authenticated', async (port) => {
	let isAuthenticated = await authenticateUser.isAuthenticated()
	port.postMessage({isAuthenticated: isAuthenticated})
})
