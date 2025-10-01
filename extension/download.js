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

	newDownloadReq: function (fileName, url, dir) {
		return {
			fileName: fileName,
			url: url,
			dir: dir,
		}
	},
	download: {
		download: async function (downloadReq) {},
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
	},
	/*
	updates: {
		callbacks: [],
		waUpdates: new WebSocket('http://localhost:1616/wa/updates'),
		onUpdate: function (callback) {
			this.callbacks.push(callback)
		},
	},
	*/
	controls: {
		pauseDownload: async function (downloadID) {},
		resumeDownload: async function (downloadID) {},
		deleteDownload: async function (downloadID) {},
	},
}
