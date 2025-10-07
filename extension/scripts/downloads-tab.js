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

	downloadsTabContainer.querySelector('.download-btn').addEventListener('click', () => {
		componentSystem.loadComponent('/downloadPopup/main.html').then((component) => {
			let dialogPopupEl = component.querySelector('dialog')
			document.body.prepend(dialogPopupEl)
			dialogPopupEl.showModal()
		})
	})

	function newDownloadedItem(data) {
		let downloadedItem = downloadsTabTemplate.querySelector('.downloaded-item').cloneNode(true)
		downloadedItem.querySelector('.download-file-name').innerText = data['file-name']
		downloadedItem.dataset.id = data.id
		downloadedItem.dataset.dateAndTime = data['date-and-time']
		let fileName = downloadedItem.querySelector('.download-file-name')
		fileName.href = `${data.dir}${navigator.platform == 'Win32' ? '\\' : '/'}${data['file-name']}`
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
			console.log(getInputValue(e).trim())
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
		let pauseResumeEl = downloadingOptions.querySelector('.pause-resume-btn')
		pauseResumeEl.dataset.status = downloader.downloadStatus(data['status'])
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

	function updateDownloadingContainerOnUpdate(downloadsTabContainer) {
		let downloadingItemContainer = downloadsTabContainer.querySelector(
			'.downloading-item-container'
		)
		let downloadingWaUpdates = new WebSocket('http://localhost:1616/download/wa/updates')
		downloadingWaUpdates.addEventListener('message', async ({data}) => {
			let json = JSON.parse(data)
			if (json.error != '' && json.error != 'deleted') {
				console.warn(json.error)
				return
			}
			let downloadItemID = json['download-id']
			console.assert(downloadItemID != undefined, 'Download ID is undefined.')

			let downloadingItemEl = downloadingItemContainer.querySelector(
				`.downloading-item[data-id="${downloadItemID}"]`
			)
			if (downloadingItemEl == null) {
				let data = await downloader.download.getDownloadingItem(downloadItemID)
				downloadingItemEl = newDownloadingItem(data)
				downloadingItemContainer.prepend(downloadingItemEl)
			}

			let pauseResumeBtn = downloadingItemEl.querySelector('.pause-resume-btn')
			if (json.error == 'deleted') {
				DeleteElementWithAnimation(downloadingItemEl)
				return
			} else if (!json['partial-content']) {
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

			function calculateProgress(length, contentLength) {
				if (contentLength == 0 || length == 0) {
					return 0
				}
				return decimalPoints((length / contentLength) * 100, 2)
			}

			downloadingItemEl.querySelector('progress').value = calculateProgress(
				json['length'],
				json['content-length']
			)
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
		})
	}

	renderSearch(downloadsTabContainer)
	renderDownloadingContainer(downloadsTabContainer)
	updateDownloadingContainerOnUpdate(downloadsTabContainer)
	renderDownloadedContainer(downloadsTabContainer)
	main.set(downloadsTabContainer)
})

function getFileExtensionNameFromFileName(filename) {
	let res = ''
	for (let i = filename.length - 1; i >= 0 && filename[i] != '.'; i--) {
		res = filename[i] + res
	}
	return res
}
