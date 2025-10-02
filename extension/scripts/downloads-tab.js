document.querySelector('.downloads-tab').addEventListener('click', () => {
	let downloadsTabTemplate = document.querySelector('#downloads-tab-template').content

	let downloadsTabContainer = downloadsTabTemplate
		.querySelector('.downloads-tab-container')
		.cloneNode(true)
	let searchBarEl = searchBar.get()
	downloadsTabContainer.prepend(searchBarEl)
	//Add copy download link and delete download item
	{
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

		let downloadedItemContainer = downloadsTabContainer.querySelector(
			'.downloaded-item-container'
		)
		downloader.download.getDownloads().then((res) => {
			console.assert(res.error == undefined, res.error)
			res['download-items'].forEach((data) => {
				downloadedItemContainer.appendChild(newDownloadedItem(data))
			})
		})
	}

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
