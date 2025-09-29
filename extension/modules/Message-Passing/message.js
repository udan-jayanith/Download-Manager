/*
 * static message contains helper functions to comminate between popup and the service worker.
 */
let message = {
	new: {
		header: function (request) {
			return {
				request: request,
				token: Math.random(),
			}
		},
		messageObject: function (header, body) {
			console.assert(typeof body == 'object', 'Expected body to be a object')
			return {
				header: header,
				body: body,
			}
		},
	},
	request: function (request, body) {
		if (body == null) {
			body = {}
		}
		return chrome.runtime.sendMessage(this.new.messageObject(this.new.header(request), body))
	},
	requests: new Map(),
	onRequest: function (request, callback) {
		this.requests.set(request, callback)
	},
}

chrome.runtime.onMessage.addListener(function (request, sender, response) {
	if (request == undefined || request.header == undefined || request.body == undefined) {
		return
	}

	let requestName = request.header.request
	let requestBody = request.body

	if (!message.requests.has(requestName)) {
		return
	}

	let callback = message.requests.get(requestName)
	console.assert(
		callback.constructor.name != 'AsyncFunction',
		callback,
		'It self cannot be async.',
		'If callback response to the request asynchronously it should return true. Default return value is false/undefined'
	)
	return callback(requestBody, response) == true
})
