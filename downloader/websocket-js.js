let wa = new WebSocket('http://localhost:1616/wa/updates')

wa.addEventListener('close', () => {
	console.log('websocket closed.')
})

wa.addEventListener('error', (err) => {
	console.log('Error occurred.', err)
})

wa.addEventListener('open', () => {
	console.log('websocket opened.')
})

wa.addEventListener('message', (e) => {
	console.log('Message received.')
	console.log(e.data)
})
