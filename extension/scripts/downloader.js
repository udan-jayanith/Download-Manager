let downloader = {
	downloadStatus: function (status) {
		switch (status) {
			case '0':
			case 0:
				return 'pending'
			case '1':
			case 1:
				return 'downloading'
			case '2':
			case 2:
				return 'complete'
			case '3':
			case 3:
				return 'paused'
		}
		return status
	},
	newDownloadReq: async function (url, filename, headers = [], filepath = '') {
		if (filepath == '') {
			let dir = await getMediaDir(getFileExtensionNameFromFileName(filename))
			filepath = dir.dir
		}

		let res = {
			url: url,
			dir: filepath,
			headers: headers,
		}
		res['file-name'] = filename
		return res
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
		getDownloadingItem: async function (id) {
			let res = await message.request('downloader.download.get-download-item', {
				id: id,
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
}
