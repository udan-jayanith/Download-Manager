let msgSocket = {
	newSocketConn: function (port) {
		let obj = {
			send: (message) => port.postMessage(message),
			onReceive: (callback) => port.onMessage.addListener(callback),
		}
		return obj
	},
	connect: function (portName) {
		console.assert(typeof portName == 'string')
		let socketPort = chrome.runtime.connect({name: portName})
		return this.newSocketConn(socketPort)
	},
}

function ConnectToSocket(portName, callback) {
	callback(msgSocket.connect(portName))
}
