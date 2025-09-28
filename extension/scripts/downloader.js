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

	getDownloads: {
		port: chrome.runtime.connect({name: 'downloader.get-downloads'}),
		get: async function (dateAndTime) {
			if (dateAndTime == null) {
				this.port.postMessage({})
			} else {
				this.port.postMessage({
					dateAndTime: dateAndTime,
				})
			}
			return new Promise((resolve, reject) => {
				this.port.onMessage.addListener((downloads) => {
					if (downloads.error != undefined) {
						reject(downloads)
						return
					}
					resolve(downloads)
				})
			})
		},
	},

	getDownloading: {
		port: chrome.runtime.connect({name: 'downloader.get-downloading'}),
		get: function () {
			this.port.postMessage({})
			return new Promise((resolve, reject) => {
				this.port.onMessage.addListener((downloading) => {
					if (downloading.error != undefined) {
						reject(downloading)
						return
					}
					resolve(downloading)
				})
			})
		},
	},

	searchDownloads: {
		port: chrome.runtime.connect({name: 'downloader.search'}),
		get: function (query) {
			console.assert(typeof query == 'string', 'Expected type of string query')
			this.port.postMessage({
				query: query,
			})
			return new Promise((resolve, reject) => {
				this.port.onMessage.addListener((results) => {
					if (results.error != null) {
						reject(results)
						return
					}
					resolve(results)
				})
			})
		},
	},
}
