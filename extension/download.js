function getDomainNames(url) {
	let domains = []
	let {hostname} = new URL(url)
	hostname = hostname
	domains.push(hostname)

	let dotCount = 0
	let simpleDomainName = ''
	for (let i = hostname.length - 1; i >= 0; i--) {
		if (hostname[i] == '.') {
			dotCount++
		}
		if (dotCount >= 2) {
			break
		}
		simpleDomainName = hostname[i] + simpleDomainName
	}

	if (simpleDomainName != hostname) {
		domains.push(simpleDomainName)
		domains.push('.' + simpleDomainName)
	}
	return domains
}

function buildCookiesValue(cookiesList) {
	let values = []
	cookiesList.forEach((cookie) => {
		values.push(`${cookie.name}=${cookie.value}`)
	})
	return values.join('; ')
}

let downloader = {
	origin: 'http://localhost:1616',
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
			try {
				let url = new URL(downloadReq.url)
				if (url.protocol == 'https:' || url.protocol == 'http:') {
					let domains = getDomainNames(url)
					let cookiesList = []
					for (let i = 0; i < domains.length; i++) {
						let domain = domains[i]

						let cookies = await chrome.cookies.getAll({
							domain: domain,
						})

						cookiesList.push(...cookies)
					}

					downloadReq.headers.push({name: 'Cookie', value: buildCookiesValue(cookiesList)})
				}
			} catch (err) {
				return {
					error: err,
				}
			}

			let res = await fetchFromDownloader('http://localhost:1616/download/download', {
				body: JSON.stringify(downloadReq),
				method: 'POST',
			})
			let json = {}
			if (res.headers.get('Content-Type') == 'application/json') {
				json = await res.json()
			}
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
		getDownloadItem: async function (downloadItemID) {
			let url = new URL(`http://localhost:1616/download/get-download-item`)
			url.searchParams.append('download-id', downloadItemID)
			let res = await fetchFromDownloader(url)
			let json = await res.json()
			return json
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

message.onRequest('downloader.download.get-download-item', ({id}, response) => {
	downloader.download.getDownloadItem(id).then((res) => {
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
let downloadUpdatesWa = new WebSocket('http://localhost:1616/download/wa/updates')
downloadUpdatesWa.addEventListener('message', ({data}) => {
	let json = JSON.parse(data)
	//10485760 == 10MB
	if (downloader.downloadStatus(json.status) != 'complete' || json['content-length'] < 10485760) {
		return
	}
	downloader.download.getDownloadItem(json['download-id']).then((res) => {
		console.assert(res.error == undefined, res.error)
		let title = `Download Complete`.trim()
		let message = `File "${res['file-name']}"`
		notify(title, message)
	})
})

let allowedDownloads = new Set()
message.onRequest('downloader.allowDownload', ({url}) => {
	allowedDownloads.add(url)
})

async function getMediaDir(extensionName) {
	let settings = await getSettings()
	let res = {
		type: 'other',
		dir: settings.othersDir,
	}

	let keys = Object.keys(settings.mediaTypes)
	for (let i = 0; i < keys.length; i++) {
		let key = keys[i]
		if (settings.mediaTypes[key][extensionName]) {
			switch (key) {
				case 'document':
					res.type = 'document'
					res.dir = settings.documentsDir
					return res
				case 'compressed':
					res.type = 'compressed'
					res.dir = settings.compressedDir
					return res
				case 'audio':
					res.type = 'audio'
					res.dir = settings.audiosDir
					return res
				case 'video':
					res.type = 'video'
					res.dir = settings.videosDir
					return res
				case 'image':
					res.type = 'image'
					res.dir = settings.imagesDir
					return res
			}
		}
	}
	return res
}

function getFileExtensionNameFromFileName(filename) {
	let res = ''
	for (let i = filename.length - 1; i >= 0 && filename[i] != '.'; i--) {
		res = filename[i] + res
	}
	return res
}

chrome.downloads.onCreated.addListener((downloadItem) => {
	if (downloadItem.state != 'in_progress') {
		return
	} else if (downloadItem.byExtensionId != undefined || allowedDownloads.has(downloadItem.url)) {
		allowedDownloads.delete(downloadItem.url)
		return
	}

	let downloadID = downloadItem.id
	chrome.downloads.cancel(downloadID)

	let filename = downloadItem.filename.trim()
	let extensionName = mediaTypeExtensionName(downloadItem.mime)

	let url = downloadItem.finalUrl
	let urlInfo = parseURL(url)
	if (urlInfo.fileName == '') {
		urlInfo.fileName = randomString(8)
	}
	if (urlInfo.extensionName != '') {
		extensionName = urlInfo.extensionName
	}
	if (filename == '') {
		filename = urlInfo.fileName + '.' + extensionName
	}

	getMediaDir(extensionName).then(({dir}) => {
		let downloadReq = downloader.newDownloadReq(url, filename, dir)
		downloader.download.download(downloadReq).then((json) => {
			if (json.error == undefined || (json.error != undefined && json.error == '')) {
				return
			}
			console.log(json.error)
		})
	})
})
