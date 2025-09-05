document.querySelector('.downloads-tab').addEventListener('click', () => {
	let downloadsTabTemplate = document.querySelector('#downloads-tab-template').content

	let downloadsTabContainer = downloadsTabTemplate
		.querySelector('.downloads-tab-container')
		.cloneNode(true)
	downloadsTabContainer.prepend(searchBar.get())

	let downloadingItemContainer = downloadsTabContainer.querySelector(
		'.downloading-item-container'
	)
	let downloadedItemContainer = downloadsTabContainer.querySelector('.downloaded-item-container')

	let downloadedItem = downloadsTabTemplate.querySelector('.downloaded-item')
	for (let i = 0; i < 10; i++) {
		downloadedItemContainer.appendChild(downloadedItem.cloneNode(true))
	}

	let downloadingItem = downloadsTabTemplate.querySelector('.downloading-item')
	downloadingItemContainer.appendChild(downloadingItem.cloneNode(true))
	main.set(downloadsTabContainer)
})
