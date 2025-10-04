document.querySelector('.settings-tab').addEventListener('click', async () => {
	let settingsContainer = document.querySelector('#settings-tab-template').content.cloneNode(true)
	let settings = await getSettings()
	//Document settings
	{
		let documentTypesInput = settingsContainer.querySelector('.document-types-input')
		documentTypesInput.value = serializeMediaTypes(settings.mediaTypes.document)
		documentTypesInput.addEventListener('input', (e) => {
			let obj = parseMediaTypes(getInputValue(e))
			settings.mediaTypes.document = obj
			updateSettings(settings)
		})

		let documentsDirInput = settingsContainer.querySelector('.documents-dir-input')
		documentsDirInput.value = settings.documentsDir
		documentsDirInput.addEventListener('input', (e) => {
			settings.documentsDir = getInputValue(e)
			updateSettings(settings)
		})
	}

	//Images settings
	{
		let typesInput = settingsContainer.querySelector('.image-types-input')
		typesInput.value = serializeMediaTypes(settings.mediaTypes.image)
		typesInput.addEventListener('input', (e) => {
			let obj = parseMediaTypes(getInputValue(e))
			settings.mediaTypes.image = obj
			updateSettings(settings)
		})

		let dirInput = settingsContainer.querySelector('.images-dir-input')
		dirInput.value = settings.imagesDir
		dirInput.addEventListener('input', (e) => {
			settings.imagesDir = getInputValue(e)
			updateSettings(settings)
		})
	}

	//Audios settings
	{
		let typesInput = settingsContainer.querySelector('.audio-types-input')
		typesInput.value = serializeMediaTypes(settings.mediaTypes.audio)
		typesInput.addEventListener('input', (e) => {
			let obj = parseMediaTypes(getInputValue(e))
			settings.mediaTypes.audio = obj
			updateSettings(settings)
		})

		let dirInput = settingsContainer.querySelector('.audios-dir-input')
		dirInput.value = settings.audiosDir
		dirInput.addEventListener('input', (e) => {
			settings.audiosDir = getInputValue(e)
			updateSettings(settings)
		})
	}

	//Videos settings
	{
		let typesInput = settingsContainer.querySelector('.video-types-input')
		typesInput.value = serializeMediaTypes(settings.mediaTypes.video)
		typesInput.addEventListener('input', (e) => {
			let obj = parseMediaTypes(getInputValue(e))
			settings.mediaTypes.video = obj
			updateSettings(settings)
		})

		let dirInput = settingsContainer.querySelector('.videos-dir-input')
		dirInput.value = settings.videosDir
		dirInput.addEventListener('input', (e) => {
			settings.videosDir = getInputValue(e)
			updateSettings(settings)
		})
	}

	//Compressed settings
	{
		let typesInput = settingsContainer.querySelector('.compressed-types-input')
		typesInput.value = serializeMediaTypes(settings.mediaTypes.compressed)
		typesInput.addEventListener('input', (e) => {
			let obj = parseMediaTypes(getInputValue(e))
			settings.mediaTypes.compressed = obj
			updateSettings(settings)
		})

		let dirInput = settingsContainer.querySelector('.compresseds-dir-input')
		dirInput.value = settings.compressedDir
		dirInput.addEventListener('input', (e) => {
			settings.compressedDir = getInputValue(e)
			updateSettings(settings)
		})
	}

	//Others
	{
		let dirInput = settingsContainer.querySelector('.others-dir-input')
		dirInput.value = settings.othersDir
		dirInput.addEventListener('input', (e) => {
			settings.othersDir = getInputValue(e)
			updateSettings(settings)
		})
	}

	//Logging
	{
		let inputEl = settingsContainer.querySelector('.log-web-requests-input')
		inputEl.value = settings.logWebRequest
		inputEl.addEventListener('input', (e) => {
			settings.logWebRequest = stringBooleanToBoolean(getInputValue(e))
			updateSettings(settings)
		})
	}

	//Syncing
	{
		let inputEl = settingsContainer.querySelector('.use-sync-settings-input')
		inputEl.value = settings.useSyncSettings
		inputEl.addEventListener('input', (e) => {
			settings.useSyncSettings = stringBooleanToBoolean(getInputValue(e))
			updateSettings(settings)
		})
	}

	main.set(settingsContainer)
})

async function getSettings() {
	return message.request('settings.get')
}

async function updateSettings(settings) {
	message.request('settings.update', settings)
}

function parseMediaTypes(inputStr) {
	let types = inputStr.split(',')
	let obj = {}
	types.forEach((type) => {
		type = type.trim()
		if (type == '') {
			return
		}
		obj[type] = true
	})
	return obj
}

function serializeMediaTypes(obj) {
	let types = Object.keys(obj)
	return types.join(', ')
}

function stringBooleanToBoolean(stringBoolean) {
	return stringBoolean.toLowerCase() == 'true'
}

async function getMediaDir(extensionName) {
	let settings = await getSettings()
	let keys = Object.keys(settings.mediaTypes)
	for (let i = 0; i < keys.length; i++) {
		let key = keys[i]
		if (settings.mediaTypes[key][extensionName]) {
			switch (key) {
				case 'document':
					return settings.documentsDir
				case 'compressed':
					return settings.compressedDir
				case 'audio':
					return settings.audiosDir
				case 'video':
					return settings.videosDir
				case 'image':
					return settings.imagesDir
			}
		}
	}
	return settings.othersDir
}
