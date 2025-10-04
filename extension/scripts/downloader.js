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
	newDownloadReq: async function (url, fileName, extensionName, headers = []) {
		return {
			fileName: `${fileName}.${extensionName}`,
			url: url,
			dir: await getMediaDir(extensionName),
			headers: headers,
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
	controls: {
		pause: async function (downloadID) {
			let res = await message.request('downloader.controls.pause', {downloadID: downloadID})
			return res
		},
		resume: async function (downloadID) {
			let res = await message.request('downloader.controls.resume', {downloadID: downloadID})
			return res
		},
		delete: async function (downloadID) {
			let res = await message.request('downloader.controls.delete', {downloadID: downloadID})
			return res
		},
	},
	updates: {
		connect: function () {
			return msgSocket.connect('downloader.downloading.updates')
		},
	},
}
