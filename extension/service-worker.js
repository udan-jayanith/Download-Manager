//This api is being used by somewhere else find it and remove it.
chrome.runtime.onConnect.ports = {}
chrome.runtime.onConnect.onPort = function (portName, callback) {
	chrome.runtime.onConnect.ports[portName] = callback
}

chrome.runtime.onConnect.addListener(function (port) {
	let callback = chrome.runtime.onConnect.ports[port.name]
	if (callback == undefined) {
		return
	}
	callback(port)
})

async function notify(title, message, iconUrl = 'http://localhost:1616/pages/favicon.png') {
	return chrome.notifications.create(null, {
		title: title,
		message: message,
		iconUrl: iconUrl,
		type: 'basic',
	})
}

importScripts(
	'./modules/Message-Passing/message.js',
	'./modules/Message-Passing/service-worker-msg-socket.js',
	'./settings.js',
	'./authentication.js',
	'./download.js',
	'./webRequests.js'
)
