let msgSocket = {
	newSocketConn: function (port) {
		let obj = {
			send: (message) => port.postMessage(message),
			onReceive: (callback) => port.onMessage.addListener(callback),
		}
		return obj
	},
	callbacks: new Map(),
	onConnect: function (portName, callback) {
		console.assert(typeof portName == 'string')
		this.callbacks.set(portName, callback)
	},
}

chrome.runtime.onConnect.addListener(function (port) {
	let callback = msgSocket.callbacks.get(port.name)
	if (callback == null) {
		return
	}
	callback(msgSocket.newSocketConn(port))
})

