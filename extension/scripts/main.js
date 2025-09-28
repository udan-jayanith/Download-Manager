let main = {
	mainEl: document.querySelector('main'),
	set: function (element) {
		this.mainEl.innerHTML = ''
		if (element != null) {
			this.mainEl.appendChild(element)
		} else {
			console.error('element is null')
		}
	},
}

let searchBar = {
	searchBarEl: document.querySelector('#search-bar-template').content,
	get: function () {
		return this.searchBarEl.cloneNode(true)
	},
}

function getInputValue(event) {
	if (event == undefined) {
		return ''
	} else if (event.target == undefined) {
		return ''
	} else if (event.target.value == undefined) {
		return ''
	}
	return event.target.value
}

async function notifyError(errObj) {
	console.assert(errObj.error != undefined)
	return chrome.notifications.create(null, {
		title: 'Error',
		message: errObj.error,
		iconUrl: 'http://localhost:1616/pages/favicon.png',
		type: 'basic',
	})
}
