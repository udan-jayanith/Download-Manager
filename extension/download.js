let downloader = {
	origin: 'http://localhost:1616',
	downloadStatus: function (status) {
		switch (status) {
			case 0:
				return 'pending'
			case 1:
				return 'downloading'
			case 2:
				return 'complete'
			case 3:
				return 'paused'
		}
		return ''
	},

	newDownloadReq: function (url, fileName, dir, headers = []) {
		let res = {
			url: url,
			dir: dir,
			headers: headers,
		}
		res['file-name'] = fileName
		return res
	},
	download: {
		download: async function (downloadReq) {
			let res = await fetchFromDownloader('http://localhost:1616/download/download', {
				body: JSON.stringify(downloadReq),
				method: 'POST',
			})
			let json = res.headers.get('Content-Type') == 'application/json' ? await res.json() : {}
			return json
		},
		getDownloads: async function (dateAndTime) {
			let url = new URL('http://localhost:1616/download/get-downloads')
			if (dateAndTime != undefined) {
				url.searchParams.append('date-and-time', dateAndTime)
			}
			let res = await fetchFromDownloader(url)
			let json = await res.json()
			return json
		},
		getDownloading: async function () {
			let res = await fetchFromDownloader('http://localhost:1616/download/get-downloading')
			let json = await res.json()
			return json
		},
		searchDownloads: async function (query) {
			let url = new URL('http://localhost:1616/download/search-downloads')
			url.searchParams.append('query', query)
			let res = await fetchFromDownloader(url)
			let json = await res.json()
			return json
		},
	},
	updates: {
		callbacks: [],
		waUpdates: new WebSocket('http://localhost:1616/download/wa/updates'),
		onUpdate: function (callback) {
			this.callbacks.push(callback)
		},
	},
	controls: {
		pauseDownload: async function (downloadID) {
			let url = new URL(`http://localhost:1616/download/pause`)
			url.searchParams.append('download-id', downloadID)
			let res = await fetchFromDownloader(url)
			return res
		},
		resumeDownload: async function (downloadID) {
			let url = new URL(`http://localhost:1616/download/resume`)
			url.searchParams.append('download-id', downloadID)
			let res = await fetchFromDownloader(url)
			return res
		},
		deleteDownload: async function (downloadID) {
			let url = new URL(`http://localhost:1616/download/delete`)
			url.searchParams.append('download-id', downloadID)
			let res = await fetchFromDownloader(url)
			return res
		},
	},
}

//Downloads
message.onRequest('downloader.download.download', (downloadReq, response) => {
	downloader.download.download(downloadReq).then((res) => {
		response(res)
	})
	return true
})

message.onRequest('downloader.download.get-downloads', ({dateAndTime}, response) => {
	downloader.download.getDownloads(dateAndTime).then((res) => {
		response(res)
	})
	return true
})

message.onRequest('downloader.download.get-downloading', (_, response) => {
	downloader.download.getDownloading().then((res) => {
		response(res)
	})
	return true
})

message.onRequest('downloader.download.search-downloads', ({query}, response) => {
	downloader.download.searchDownloads(query).then((res) => {
		response(res)
	})
	return true
})

//Controls
message.onRequest('downloader.controls.pause', ({downloadID}, response) => {
	downloader.controls.pauseDownload(downloadID).then((err) => {
		response(err)
	})
	return true
})

message.onRequest('downloader.controls.resume', ({downloadID}, response) => {
	downloader.controls.resumeDownload(downloadID).then((err) => {
		response(err)
	})
	return true
})

message.onRequest('downloader.controls.delete', ({downloadID}, response) => {
	downloader.controls.deleteDownload(downloadID).then((err) => {
		response(err)
	})
	return true
})

//Downloading updates (wa updates)
downloader.updates.waUpdates.addEventListener('message', ({data}) => {
	downloader.updates.callbacks.forEach((callback) => {
		callback(data)
	})
})

msgSocket.onConnect('downloader.downloading.updates', (conn) => {
	downloader.updates.onUpdate((data) => {
		conn.send(data)
	})
})
