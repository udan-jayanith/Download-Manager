# Message passing

This library is created to simplify communicate between chrome extension popup and the service-worker more intuitively and uses chrome message passing api in underlying layer.

## One way connections

One way connections allows you to make a request to the service worker. Service worker also can send a response to the request just like in HTTP.
To handle requests and to requests import `message.js`.

### Popup

```js
message.request('service-worker-1').then((res) => {
	console.log('response received from the service worker-1')
})
```

`message.request` function optionally takes a object and send it with the request.
`res` is the response from the service worker. In this case `{}` because service worker doesn't return anything.
`message.onRequest` function shouldn't be used in the popup.

### Service worker

`onRequest` takes a port name and a callback function. When a request is made to `service-worker-1` port callback get called by passing the request `body` and a `response`.
Callback itself cannot be asynchronous.
`response` takes a value of type object and response to the request. If response get called asynchronously callback should return true.

```js
message.onRequest('service-worker-1', (body, response) => {
	console.log('service-worker-1')
})
```

Here service worker logs `service-worker-1`(Request name) when a request is made to `service-worker-1`.
`body` is the request body. You can optionally send a response to the request by retuning a value of type object.
`message.request` function shouldn't be used in the service-worker.

## Long lived connections

Long lived connections allows to send and receive messages to both ends (Service worker and popup).

### Service worker

Import `service-worker-msg-socket.js` to use long lived connections in the service worker.

```js
msgSocket.onConnect('socket', (conn) => {
	conn.send({
		msg: 'hello',
	})
	conn.onReceive((obj) => {
		console.assert(obj.msg != undefined)
		if (obj.msg == 'welcome') {
			conn.send({msg: 'thanks'})
		} else {
			console.log(obj.msg)
		}
	})
})
```

Here when popup connected to `socket` socket port service worker sends a message to popup and then starts listening incoming messages.

### Popup

Import `content-script-msg-socket.js` to connect to msgSockets in popup.

```js
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
```

Here popup connects to `socket` socket port and starts listening to incoming messages. Popup `conn` also has the same methods in service worker conn.
