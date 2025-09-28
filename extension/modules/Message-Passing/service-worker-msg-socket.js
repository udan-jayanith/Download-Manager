 let msgSocket = {
	callbacks: new Map(),
	onConnect: function (portName, callback) {
		console.assert(typeof portName == 'string')
		this.callbacks.set(portName, callback)
	},
}

function newSocketConn(port) {
	let obj = {
		send: (message) => port.postMessage(message),
		onReceive: (callback) => port.onMessage.addListener(callback),
	}
	return obj
}

chrome.runtime.onConnect.addListener(function (port) {
	let callback = msgSocket.callbacks.get(port.name)
	if (callback == null) {
		return
	}
	callback(newSocketConn(port))
})
