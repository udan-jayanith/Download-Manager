document.querySelector('.resources-tab').addEventListener('click', () => {
	setResourcesTab()
})

async function setResourcesTab() {
	let settings = await getSettings()
	let webRequestsMap = new Map()

	function render(resourcesContainer, webRequests) {
		resourcesContainer.querySelectorAll('.resources-item').forEach((el) => {
			el.remove()
		})
		webRequestsMap.clear()

		let resourcesItem = resourcesTabTemplate.querySelector('.resources-item')
		webRequests.forEach((el) => {
			webRequestsMap.set(el.requestId, el)
			let item = resourcesItem.cloneNode(true)
			let filenameEl = item.querySelector('.resources-file-name')
			filenameEl.innerText = el.fileName + '.' + el.extensionName
			filenameEl.dataset.requestId = el.requestId
			item.title = el.extensionName
			item
				.querySelector('.fa-solid')
				.classList.add(getItemIconClassName(settings.mediaTypes, el.extensionName))
			item.dataset.requestId = el.requestId
			item.querySelector('.copy-download-link-btn').dataset.url = el.url

			resourcesContainer.appendChild(item)
		})
	}

	function getItemIconClassName(mediaTypes, extensionName) {
		let keys = Object.keys(mediaTypes)
		for (let i = 0; i < keys.length; i++) {
			if (mediaTypes[keys[i]][extensionName] == undefined) {
				continue
			}
			switch (keys[i]) {
				case 'document':
					return 'fa-file'
				case 'compressed':
					return 'fa-file-zipper'
				case 'audio':
					return 'fa-file-audio'
				case 'video':
					return 'fa-file-video'
				case 'image':
					return 'fa-image'
			}
		}
		return 'fa-circle-question'
	}

	let resourcesTabTemplate = document.querySelector('#resources-tab-template').content
	let resourcesContainer = resourcesTabTemplate
		.querySelector('.resources-container')
		.cloneNode(true)

	EventDelegation(resourcesContainer, '.copy-download-link-btn', 'click', (e) => {
		navigator.clipboard.writeText(e.dataset.url)
	})

	EventDelegation(resourcesContainer, '.resources-file-name', 'click', async (target) => {
		console.assert(target != undefined, 'target is undefined.')
		let requestId = target.dataset.requestId
		console.assert(requestId != undefined, 'requestId is undefined.')

		let webRequest = webRequestsMap.get(requestId)
		console.assert(webRequest != undefined, 'webRequest is not found in the webRequestsMap.')
		let downloadReq = await downloader.newDownloadReq(
			webRequest.url,
			webRequest.fileName,
			webRequest.extensionName,
			webRequest.headers
		)
		let res = await downloader.download.download(downloadReq)
		console.assert(res.error == undefined, res.error)
	})

	let searchBarEl = searchBar.get()
	let searchBarInput = searchBarEl.querySelector('input')
	searchBarInput.addEventListener('keyup', (e) => {
		resourcesSearch(getInputValue(e), selectedFilter.dataset.value).then((webRequests) => {
			render(resourcesContainer, webRequests)
		})
	})
	resourcesContainer.prepend(searchBarEl)

	let filterTagsContainer = resourcesContainer.querySelector('.filter-tags-container')
	let selectedFilter = filterTagsContainer.querySelector('.media-tag')
	filterTagsContainer.addEventListener('click', (e) => {
		let tagEl = e.target.closest('.tag')
		if (tagEl == null) {
			return
		}
		selectedFilter = tagEl
		filterTagsContainer.querySelectorAll('.tag').forEach((el) => {
			el.classList.remove('selected-tag')
		})
		selectedFilter.classList.add('selected-tag')
		resourcesSearch(searchBarInput.value, selectedFilter.dataset.value).then((webRequests) => {
			render(resourcesContainer, webRequests)
		})
	})
	selectedFilter.classList.add('selected-tag')

	resourcesSearch('', selectedFilter.dataset.value).then((webRequests) => {
		render(resourcesContainer, webRequests)
	})

	main.set(resourcesContainer)
}

async function getWebRequests() {
	let res = await message.request('webRequests')
	return res.webRequest
}

async function resourcesSearch(query, selectedFilter) {
	let webRequests = await getWebRequests()
	query = query.toLowerCase()
	let mediaTypes = (await getSettings()).mediaTypes

	return webRequests.filter((el) => {
		switch (selectedFilter) {
			case 'media':
				let keys = Object.keys(mediaTypes)
				let contained = false
				for (let i = 0; i < keys.length; i++) {
					if (mediaTypes[keys[i]][el.extensionName]) {
						contained = true
						break
					}
				}
				if (!contained) {
					return false
				}
				break
			case 'all':
				break
			default:
				if (mediaTypes[selectedFilter][el.extensionName] == undefined) {
					return false
				}
		}

		if (query.trim() == '') {
			return true
		}

		for (let value in el) {
			let str = el[value]
			if (typeof str != 'string') {
				str = String(str)
			}
			str = str.toLowerCase()
			if (str.includes(query)) {
				return true
			}
		}
		return false
	})
}

document.querySelector('.resources-tab').classList.add('selected-nav-item')
setResourcesTab()
