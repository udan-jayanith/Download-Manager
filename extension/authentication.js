let authenticateUser = {
	saveToken: async function (token) {
		if (!(await this.isValidToken(token))) {
			return {
				error: 'Invalid token.',
			}
		}

		let settings = await getSettings()
		if (settings.cookies == undefined) {
			settings.cookies = {}
		}
		settings.cookies.token = token
		updateSettings(settings)
		return {}
	},
	getToken: async function () {
		let settings = await getSettings()
		return settings.cookies.token
	},
	isValidToken: async function (token) {
		let formdata = new FormData()
		formdata.append('token', token)
		let res = await fetchFromDownloader('http://localhost:1616/token/is-valid', {
			body: formdata,
			method: 'POST',
		})
		let json = await res.json()
		if (json.error != undefined) {
			console.error(json.error)
			return false
		}
		return json['is-valid']
	},
}

async function addAuthorization(headers) {
	headers.append('Authorization', `Bearer ${authenticateUser.getToken()}`)
	return headers
}

function fetchFromDownloader(resource, options = {}) {
	if (options.headers == null) {
		options.headers = new Headers()
	} else {
		options.headers.delete('Authorization')
	}
	options.headers = addAuthorization(options.headers)
	return fetch(resource, options)
}

message.onRequest('authentication.save-token', ({token}, response) => {
	authenticateUser.saveToken(token).then((res) => {
		response(res)
	})
	return true
})

message.onRequest('authentication.is-authenticated', (_, response) => {
	authenticateUser.getToken().then(async (token) => {
		response({isAuthenticated: await authenticateUser.isValidToken(token)})
	})
	return true
})
