document.querySelector('.downloads-tab').addEventListener('click', () => {
	let downloadsTabTemplate = document.querySelector('#downloads-tab-template').content

	let downloadsTabContainer = downloadsTabTemplate
		.querySelector('.downloads-tab-container')
		.cloneNode(true)

	//Search bear
	{
		let searchBarEl = searchBar.get()
		downloadsTabContainer.prepend(searchBarEl)
	}

	EventDelegation(downloadsTabContainer, '.copy-download-link-btn', 'click', (el) => {
		navigator.clipboard.writeText(el.dataset.url)
	})

	EventDelegation(downloadsTabContainer, '.delete-download-item-btn', 'click', (el) => {
		console.assert(el.dataset.id != null, 'ID is null')
		let downloadItem = el.closest('.downloaded-item')
		downloader.controls.delete(el.dataset.id).then((res) => {
			if (res.error != undefined) {
				notifyError(res)
				return
			}
			if (downloadItem != null) {
				downloadItem.remove()
			}
		})
	})

	function newDownloadedItem(data) {
		let downloadedItem = downloadsTabTemplate.querySelector('.downloaded-item').cloneNode(true)
		downloadedItem.querySelector('.download-file-name').innerText = data['file-name']
		downloadedItem.dataset.id = data.id
		downloadedItem.dataset.dateAndTime = data['date-and-time']

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

	function renderDownloadedList() {
		let downloadedItemContainer = downloadsTabContainer.querySelector(
			'.downloaded-item-container'
		)
		downloadedItemContainer.innerHTML = null
		downloader.download.getDownloads().then((res) => {
			console.assert(res.error == undefined, res.error)
			res['download-items'].forEach((data) => {
				downloadedItemContainer.appendChild(newDownloadedItem(data))
			})
		})
	}
	renderDownloadedList()

	//downloadingItem
	{
		let downloadingItemContainer = downloadsTabContainer.querySelector(
			'.downloading-item-container'
		)
		let downloadingItem = downloadsTabTemplate.querySelector('.downloading-item')
		downloadingItemContainer.appendChild(downloadingItem.cloneNode(true))
	}
	main.set(downloadsTabContainer)
})
