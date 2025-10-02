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

	download: {
		download: async function (downloadReq) {
			console.assert(
				typeof downloadReq == 'object',
				'downloadReq must be a downloadReq of type object'
			)
			let res = await message.request('downloader.download.download', downloadReq)
			return res
		},
		getDownloads: async function (dateAndTime) {
			let res = await message.request('downloader.download.get-downloads', {
				dateAndTime: dateAndTime,
			})
			return res
		},
		getDownloading: async function () {
			let res = await message.request(`downloader.download.get-downloading`)
			return res
		},
		searchDownload: async function (query) {
			let res = await message.request('downloader.download.search-downloads', {
				query: query,
			})
			return res
		},
	},
	controls: {},
	updates: {},
	/*
	updates: {
		port: chrome.runtime.connect({name: 'downloader.waUpdates'}),
		callbacks: [],
		onUpdate: function (callback) {
			this.callbacks.push(callback)
		},
	},
	*/
}
