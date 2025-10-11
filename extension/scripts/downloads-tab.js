document.querySelector('.downloads-tab').addEventListener('click', () => {
	let downloadsTabTemplate = document.querySelector('#downloads-tab-template').content

	let downloadsTabContainer = downloadsTabTemplate
		.querySelector('.downloads-tab-container')
		.cloneNode(true)

	EventDelegation(downloadsTabContainer, '.copy-download-link-btn', 'click', (el) => {
		navigator.clipboard.writeText(el.dataset.url)
	})

	EventDelegation(downloadsTabContainer, '.delete-download-item-btn', 'click', (el) => {
		console.assert(el.dataset.id != null, 'ID is null')
		let downloadTabContainer = el.closest('.downloads-tab-container')
		let downloadItems = downloadTabContainer.querySelectorAll(
			`.downloaded-item[data-id='${el.dataset.id}'], .downloading-item[data-id='${el.dataset.id}']`
		)
		downloader.controls.delete(el.dataset.id).then((res) => {
			if (res.error != undefined) {
				notifyError(res)
				return
			}
			downloadItems.forEach((el) => {
				DeleteElementWithAnimation(el)
			})
		})
	})

	EventDelegation(
		downloadsTabContainer,
		'.downloaded-item .download-file-name',
		'click',
		(el) => {
			let location = el.dataset.location
			chrome.tabs
				.query({
					active: true,
					currentWindow: true,
				})
				.then((res) => {
					if (res == undefined || res.length <= 0) {
						return
					}
					let activeTab = res[0]
					let url = new URL(activeTab.url)
					if (url.protocol == 'file:') {
						chrome.tabs.update(activeTab.id, {
							active: true,
							url: location,
						})
					} else {
						window.open(location, '_blank')
					}
				})
		}
	)

	downloadsTabContainer.querySelector('.download-btn').addEventListener('click', () => {
		componentSystem.loadComponent('/downloadPopup/main.html').then((component) => {
			let dialogPopupEl = component.querySelector('dialog')
			document.body.prepend(dialogPopupEl)

			let inputEl = downloadsTabContainer.querySelector('.download-input-container input')
			handleDownloadDialogPopup(dialogPopupEl, inputEl.value)
			inputEl.value = ''
		})
	})

	function newDownloadedItem(data) {
		let downloadedItem = downloadsTabTemplate.querySelector('.downloaded-item').cloneNode(true)
		downloadedItem.querySelector('.download-file-name').innerText = data['file-name']
		downloadedItem.dataset.id = data.id
		downloadedItem.dataset.dateAndTime = data['date-and-time']
		let fileName = downloadedItem.querySelector('.download-file-name')
		fileName.dataset.location = `${data.dir}${navigator.platform == 'Win32' ? '\\' : '/'}${
			data['file-name']
		}`
		fileName.title = getFileExtensionNameFromFileName(data['file-name'])

		let downloadItemOptions = downloadedItem.querySelector('.download-item-options')
		downloadItemOptions.querySelector('.copy-download-link-btn').dataset.url = data.url
		downloadItemOptions.querySelector('.delete-download-item-btn').dataset.id = data.id

		let fileSize = byte(Number(data['content-length'])).get()
		downloadedItem.querySelector('.file-size').innerText =
			decimalPoints(fileSize.data, 2) + ' ' + fileSize.unit
		downloadedItem.querySelector('.date-and-time').innerText = dateAndTimeAgo(
			data['date-and-time']
		)

		return downloadedItem
	}

	function renderSearch(downloadsTabContainer) {
		let searchBarEl = searchBar.get()
		let searchBarInputEl = searchBarEl.querySelector('input')

		let searchResultsContainerEl = downloadsTabContainer.querySelector(
			'.search-results-container'
		)
		hideEl(searchResultsContainerEl)

		let rendering = false
		searchBarInputEl.addEventListener('input', (e) => {
			if (getInputValue(e).trim() == '') {
				hideEl(searchResultsContainerEl)
				searchResultsContainerEl.innerHTML = null
				return
			} else if (rendering) {
				return
			} else {
				rendering = true
			}

			downloader.download.searchDownload(getInputValue(e)).then((res) => {
				console.assert(res.error == undefined, res.error)
				let searchResults = res['search-results']
				if (searchResults == undefined) {
					return
				}
				showEl(searchResultsContainerEl)
				searchResultsContainerEl.innerHTML = null
				searchResults.forEach((data) => {
					searchResultsContainerEl.appendChild(newDownloadedItem(data))
				})
				rendering = false
			})
		})

		downloadsTabContainer.prepend(searchBarEl)
	}

	function renderDownloadedContainer(downloadsTabContainer) {
		function renderDownloadedList(list) {
			let downloadedItemContainer = downloadsTabContainer.querySelector(
				'.downloaded-item-container'
			)
			list.forEach((data) => {
				downloadedItemContainer.appendChild(newDownloadedItem(data))
			})
		}

		let lastDateAndTime = undefined
		let rendering = false
		function renderDownloaded() {
			if (rendering) {
				return
			}
			rendering = true
			downloader.download.getDownloads(lastDateAndTime).then((res) => {
				console.assert(res.error == undefined, res.error)
				let list = res['download-items']
				if (list.length <= 0) {
					return
				}
				lastDateAndTime = list[list.length - 1]['date-and-time']
				renderDownloadedList(list)
				rendering = false
			})
		}

		downloadsTabContainer.addEventListener('scroll', ({target}) => {
			let scrollBottom = target.offsetHeight + target.scrollTop
			//target.scrollHeight means full scrollable height.
			if (scrollBottom + (scrollBottom / 100) * 10 >= target.scrollHeight) {
				renderDownloaded()
			}
		})

		renderDownloaded()
	}

	function newDownloadingItem(data) {
		let downloadingItem = downloadsTabTemplate.querySelector('.downloading-item').cloneNode(true)
		downloadingItem.dataset.id = data.id
		downloadingItem.dataset.partialContent = data['partial-content']
		downloadingItem.querySelector('.download-file-name').innerText = data['file-name']

		let downloadingOptions = downloadingItem.querySelector('.download-item-options')
		let downloadStatus = downloader.downloadStatus(data['status'])

		let pauseResumeEl = downloadingOptions.querySelector('.pause-resume-btn')
		let iconEl = pauseResumeEl.querySelector('i')
		if (downloadStatus == 'paused') {
			iconEl.classList.remove('fa-circle-pause')
			iconEl.classList.add('fa-play')
		} else {
			iconEl.classList.remove('fa-play')
			iconEl.classList.add('fa-circle-pause')
		}
		pauseResumeEl.dataset.status = downloadStatus
		downloadingOptions.querySelector('.copy-download-link-btn').dataset.url = data.url
		downloadingOptions.querySelector('.delete-download-item-btn').dataset.id = data.id

		return downloadingItem
	}

	function renderDownloadingContainer(downloadsTabContainer) {
		let downloadingItemContainer = downloadsTabContainer.querySelector(
			'.downloading-item-container'
		)
		downloadingItemContainer.innerHTML = null

		downloader.download.getDownloading().then((res) => {
			console.assert(res.error == undefined, res.error)
			let list = res['downloading-items']
			list.forEach((data) => {
				downloadingItemContainer.appendChild(newDownloadingItem(data))
			})
		})
	}

	async function updateDownloadingItem(downloadingItemContainer, json) {
		let downloadItemID = json['download-id']
		let downloadingItemEl = downloadingItemContainer.querySelector(
			`.downloading-item[data-id="${downloadItemID}"]`
		)
		if (downloadingItemEl == null) {
			let data = await downloader.download.getDownloadingItem(downloadItemID)
			downloadingItemEl = newDownloadingItem(data)
			downloadingItemContainer.prepend(downloadingItemEl)
		}

		let pauseResumeBtn = downloadingItemEl.querySelector('.pause-resume-btn')
		if (!json['partial-content']) {
			hideEl(pauseResumeBtn)
		} else if (json['partial-content']) {
			showEl(pauseResumeBtn)
		}

		let downloadStatus = downloader.downloadStatus(json['status'])
		pauseResumeBtn.dataset.status = downloadStatus
		let iconEl = pauseResumeBtn.querySelector('i')
		if (downloadStatus == 'complete') {
			DeleteElementWithAnimation(downloadingItemEl)
			let data = await downloader.download.getDownloadingItem(downloadItemID)
			let downloadedItem = newDownloadedItem(data)
			let downloadedItemContainer = downloadsTabContainer.querySelector(
				'.downloaded-item-container'
			)
			downloadedItemContainer.prepend(downloadedItem)
			return
		} else if (downloadStatus == 'paused') {
			iconEl.classList.remove('fa-circle-pause')
			iconEl.classList.add('fa-play')
		} else {
			iconEl.classList.remove('fa-play')
			iconEl.classList.add('fa-circle-pause')
		}
		downloadingItemEl.querySelector('progress').value = (function (length, contentLength) {
			if (contentLength == 0 || length == 0) {
				return 0
			}
			return decimalPoints((length / contentLength) * 100, 2)
		})(json['length'], json['content-length'])

		let downloadBottom = downloadingItemEl.querySelector('.download-bottom')
		let contentLength = byte(json['content-length']).get()
		let length = byte(json['length']).get()
		downloadBottom.querySelector('.completion').innerText = `${
			decimalPoints(length.data, 2) + ' ' + length.unit
		} of ${decimalPoints(contentLength.data, 2) + ' ' + contentLength.unit}`

		let estimatedTime = seconds(json['estimated-time'])
		downloadBottom.querySelector('.estimated-time').innerText = `${decimalPoints(
			estimatedTime.count,
			2
		)} ${estimatedTime.unit}`
		let unit = byte(json['bps']).get()
		downloadBottom.querySelector('.download-speed').innerText = (function () {
			if (unit.unit == 'Byte') {
				unit.unit = 'BPS'
			} else {
				unit.unit = unit.unit + 'PS'
			}
			return decimalPoints(unit.data, 2) + ' ' + unit.unit
		})()
	}

	function updateDownloadingContainerOnUpdate(downloadsTabContainer) {
		let downloadingItemContainer = downloadsTabContainer.querySelector(
			'.downloading-item-container'
		)

		let downloadingWaUpdates = new WebSocket('http://localhost:1616/download/wa/updates')
		downloadingWaUpdates.addEventListener('message', async ({data}) => {
			let json = JSON.parse(data)
			if (json.error != '' && json.error != 'deleted') {
				console.log(json.error)
				return
			}
			let downloadItemID = json['download-id']
			console.assert(downloadItemID != undefined, 'Download ID is undefined.')

			if (json.error == 'deleted') {
				let downloadItemEl = downloadsTabContainer.querySelector(
					`.downloading-item-container .downloading-item[data-id="${downloadItemID}"], .downloaded-item-container .downloaded-item[data-id="${downloadItemID}"]`
				)
				if (downloadItemEl == null) {
					return
				}
				DeleteElementWithAnimation(downloadItemEl)
			} else {
				updateDownloadingItem(downloadingItemContainer, json)
			}
		})
	}

	let downloadingItemContainer = downloadsTabContainer.querySelector(
		'.downloading-item-container'
	)
	EventDelegation(downloadingItemContainer, '.pause-resume-btn', 'click', (el) => {
		let downloadingItem = el.closest('.downloading-item')
		console.assert(downloadingItem != null, 'DownloadingItem el is null')
		if (!downloadingItem.dataset.partialContent) {
			let msg = 'Play/Pause is not supported for this download.'
			alert(msg)
			return
		}
		let status = downloader.downloadStatus(el.dataset.status)
		let downloadItemID = Number(downloadingItem.dataset.id)
		if (status == 'paused') {
			downloader.controls.resume(downloadItemID).then((json) => {
				console.assert(json.error == undefined, json.error)
			})
		} else {
			downloader.controls.pause(downloadItemID).then((json) => {
				console.assert(json.error == undefined, json.error)
			})
		}
	})

	renderSearch(downloadsTabContainer)
	renderDownloadingContainer(downloadsTabContainer)
	updateDownloadingContainerOnUpdate(downloadsTabContainer)
	renderDownloadedContainer(downloadsTabContainer)
	main.set(downloadsTabContainer)
})

function getExtensionNameFromURL(url) {
	try {
		let {pathname} = new URL(url)
		return getFileExtensionNameFromFileName(pathname)
	} catch (err) {
		return ''
	}
}

function getFilename(pathname) {
	let res = ''
	for (let i = pathname.length - 1; i >= 0 && pathname[i] != '/'; i--) {
		res = pathname[i] + res
	}
	return res
}

async function handleDownloadDialogPopup(el, url = '') {
	el.showModal()
	let warnEl = el.querySelector('.warn')
	hideEl(warnEl)

	let urlInputEl = el.querySelector('.download-url')
	urlInputEl.value = url

	let saveFilenameEl = el.querySelector('.save-file-name')
	function getSaveFilename(url) {
		try {
			let {pathname} = new URL(url)
			return getFilename(pathname)
		} catch (err) {
			return ''
		}
	}
	saveFilenameEl.value = getSaveFilename(url)

	function closeDialog(dialogEl) {
		dialogEl.close()
		dialogEl.remove()
	}

	el.querySelector('.done-btn').addEventListener('click', async () => {
		let url = urlInputEl.value.trim()
		if (url == '') {
			warnEl.innerText = 'URL cannot be empty'
			showEl(warnEl)
			return
		}

		let saveFilename =
			saveFilenameEl.value.trim() != '' ? saveFilenameEl.value.trim() : getSaveFilename(url)
		if (saveFilename == '') {
			warnEl.innerText = 'Save file name is empty'
			showEl(warnEl)
			return
		}

		try {
			new URL(url)

			let downloadReq = await downloader.newDownloadReq(url, saveFilename)
			downloader.download.download(downloadReq).then(() => {
				closeDialog(el)
			})
		} catch (err) {
			warnEl.innerText = err
			showEl(warnEl)
		}
	})

	el.querySelector('.download-with-chrome').addEventListener('click', () => {
		let url = urlInputEl.value.trim()
		if (url == '') {
			warnEl.innerText = 'URL cannot be empty'
			showEl(warnEl)
			return
		}

		let saveFilename = saveFilenameEl.value.trim()
		try {
			new URL(url)

			let downloadReq = {
				url: url,
			}
			if (saveFilename != '') {
				downloadReq.filename = saveFilename
			}

			message.request('downloader.allowDownload', downloadReq)
			chrome.downloads.download(downloadReq).then(() => {
				closeDialog(el)
			})
		} catch (err) {
			warnEl.innerText = err
			showEl(warnEl)
		}
	})

	el.querySelector('.cancel-btn').addEventListener('click', () => {
		closeDialog(el)
	})
}
