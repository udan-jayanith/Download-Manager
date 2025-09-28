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

chrome.runtime.onMessage.addListener(function (request, sender, sendResponse) {
	if (request == undefined || request.header == undefined || request.body == undefined) {
		return
	}

	let requestName = request.header.request
	let requestBody = request.body

	if (!message.requests.has(requestName)) {
		return
	}

	let res = message.requests.get(requestName)(requestBody)
	if (res == undefined) {
		res = {}
	}
	console.assert(typeof res == 'object')
	sendResponse(res)
})
