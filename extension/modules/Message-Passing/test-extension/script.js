message.request('service-worker-1').then((res) => {
	console.log('response received from the service worker-1')
	console.log(res)
})

message.request('service-worker-2').then((res) => {
	console.log('response received from the service worker-2')
})

let conn = msgSocket.connect('socket')
conn.onReceive((obj) => {
	console.assert(obj.msg != undefined)
	if (obj.msg == 'hello') {
		conn.send({
			msg: 'welcome',
		})
	} else {
		console.log(obj.msg)
	}
})

/*
ConnectToSocket('socket', (conn) => {
	conn.onReceive((obj) => {
		console.assert(obj.msg != undefined)
		if (obj.msg == 'hello') {
			conn.send({
				msg: 'welcome',
			})
		} else {
			console.log(obj.msg)
		}
	})
})

*/