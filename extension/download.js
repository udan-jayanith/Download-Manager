async function addAuthorization(headers) {
	let settings = await getSettings()

	headers.append('Authorization', `Bearer ${settings.cookies.token}`)
	return headers
}

let downloader = {
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

	newDownloadReq: function (fileName, url, dir) {
		return {
			fileName: fileName,
			url: url,
			dir: dir,
		}
	},
	download: async function (downloadReq) {},
	/*
	updates: {
		callbacks: [],
		waUpdates: new WebSocket('http://localhost:1616/wa/updates'),
		onUpdate: function (callback) {
			this.callbacks.push(callback)
		},
	},
	*/
	getDownloads: async function (dateAndTime) {
		let url = new URL('http://localhost:1616//get-downloads')
		if (dateAndTime != undefined) {
			url.searchParams().append('date-and-time', dateAndTime)
		}

		let headers = new Headers()
		headers = await addAuthorization(headers)
		let res = await fetch(url.href, {
			headers: headers,
		})
		let json = await res.json()
		return json
	},
	getDownloading: async function () {
		let url = new URL('http://localhost:1616/get-downloading')
		let headers = await addAuthorization(new Headers())
		let res = await fetch(url, {
			headers: headers,
		})
		let json = await res.json()
		return json
	},
	searchDownloads: async function (query) {
		let url = new URL('http://localhost:1616/search-downloads')
		url.searchParams.append('query', query)
		let headers = await addAuthorization(new Headers())

		let res = await fetch(url, {
			headers: headers,
		})
		let json = await res.json()
		return json
	},

	pauseDownload: async function (downloadID) {},
	resumeDownload: async function (downloadID) {},
	deleteDownload: async function (downloadID) {},
}

chrome.runtime.onConnect.onPort('downloader.get-downloads', (port) => {
	port.onMessage.addListener(async (obj) => {
		let downloads = null
		let dateAndTime = obj == undefined ? undefined : obj.dateAndTime
		if (dateAndTime != undefined && dateAndTime.trim() != '') {
			downloads = await downloader.getDownloads()
		} else {
			downloads = await downloader.getDownloads(dateAndTime)
		}
		port.postMessage(downloads)
	})
})

chrome.runtime.onConnect.onPort('downloader.get-downloading', (port) => {
	port.onMessage.addListener(async () => {
		let downloading = await downloader.getDownloading()
		port.postMessage(downloading)
	})
})

chrome.runtime.onConnect.onPort('downloader.search', (port) => {
	port.onMessage.addListener(async (obj) => {
		console.assert(obj.query != undefined && typeof obj.query == 'string', 'Unexpected query.')
		let results = await downloader.searchDownloads(obj.query)
		port.postMessage(results)
	})
})

/*
downloader.updates.addEventListener('message', (e) => {
	downloader.updates.callbacks((callback) => {
		callback(e.data)
	})
})
*/

/*
chrome.runtime.onConnect.onPort('downloader.waUpdates', (port) => {
	downloader.updates.onUpdate((update) => {
		port.postMessage(update)
	})
})

*/
