document.querySelector('.resources-tab').addEventListener('click', () => {
	setResourcesTab()
})

async function setResourcesTab() {
	function render(resourcesContainer, webRequests) {
		resourcesContainer.querySelectorAll('.resources-item').forEach((el) => {
			el.remove()
		})
		let resourcesItem = resourcesTabTemplate.querySelector('.resources-item')
		webRequests.forEach((el) => {
			let item = resourcesItem.cloneNode(true)
			item.querySelector('.resources-file-name').innerText = el.fileName + '.' + el.extensionName
			resourcesContainer.appendChild(item)
		})
	}

	let resourcesTabTemplate = document.querySelector('#resources-tab-template').content
	let resourcesContainer = resourcesTabTemplate
		.querySelector('.resources-container')
		.cloneNode(true)

	let searchBarEl = searchBar.get()
	searchBarEl.querySelector('input').addEventListener('keyup', (e) => {
		resourcesSearch(getInputValue(e)).then((webRequests) => {
			render(resourcesContainer, webRequests)
		})
	})
	resourcesContainer.appendChild(searchBarEl)

	let webRequests = await getWebRequests()
	render(resourcesContainer, webRequests)

	main.set(resourcesContainer)
}

async function getWebRequests() {
	let webRequestPort = chrome.runtime.connect({ name: 'webRequests' })
	let webRequests = await new Promise((resolve) => {
		webRequestPort.onMessage.addListener((settings) => {
			resolve(settings.webRequest)
		})
	})
	return webRequests
}

async function resourcesSearch(query) {
	let webRequests = await getWebRequests()
	if (query.trim() == '') {
		return webRequests
	}
	query = query.toLowerCase()

	return webRequests.filter((el) => {
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

setResourcesTab()
