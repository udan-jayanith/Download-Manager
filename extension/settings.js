let settings = {
	cookies: {
		token: '',
	},
	mediaTypes: {},
	logWebRequest: false,
	useSyncSettings: true,
	documentsDir: '/Documents',
	videosDir: '/Videos',
	audiosDir: '/Music',
	imagesDir: '/Pictures',
	compressedDir: '/Compressed',
	othersDir: '/Downloads',
}

settings.mediaTypes = {
	document: {
		doc: true,
		docx: true,
		odt: true,
		wpd: true,
		pages: true,
		ppt: true,
		pptx: true,
		key: true,
		pdf: true,
		drw: true,
		dwg: true,
	},
	compressed: {
		exe: true,
		app: true,
		bin: true,
		'7z': true,
		ace: true,
		arj: true,
		bz2: true,
		cab: true,
		gz: true,
		iso: true,
		zip: true,
	},
	audio: {
		wav: true,
		aac: true,
		wma: true,
		mp3: true,
		aiff: true,
		m4a: true,
		wma: true,
	},
	video: {
		mov: true,
		avi: true,
		wmv: true,
		mp4: true,
		webm: true,
		mkv: true,
		ogg: true,
		ts: true,
	},
	image: {
		apng: true,
		png: true,
		avif: true,
		gif: true,
		jpg: true,
		jpeg: true,
		pjpeg: true,
		svg: true,
		webp: true,
		bmp: true,
		ico: true,
	},
}

getSettings().then((res) => {
	if (res == undefined || Object.keys(res).length == 0) {
		updateSettings(settings)
		return
	}
	settings = res
})

async function getSettings() {
	let syncSettings = await chrome.storage.sync.get(['settings'])
	if (syncSettings.useSyncSettings) {
		return syncSettings.settings
	}
	let localSettings = await chrome.storage.local.get(['settings'])
	return localSettings.settings
}

async function updateSettings(settingsObj) {
	await chrome.storage.sync.set({settings: settingsObj})
	await chrome.storage.local.set({settings: settingsObj})
	settings = settingsObj
}

message.onRequest('settings.get', (_, response) => {
	getSettings().then((settings) => {
		response(settings)
	})
	return true
})

chrome.runtime.onConnect.onPort('update-settings', (port) => {
	port.onMessage.addListener((settings) => {
		updateSettings(settings)
	})
})
