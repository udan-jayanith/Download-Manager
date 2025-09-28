function ConnectToSocket(port, callback) {
	console.assert(typeof port == 'string')
	let socketPort = chrome.runtime.connect({name: port})
	callback(newSocketConn(socketPort))
}

function newSocketConn(port) {
	let obj = {
		send: (message) => port.postMessage(message),
		onReceive: (callback) => port.onMessage.addListener(callback),
	}
	return obj
}
