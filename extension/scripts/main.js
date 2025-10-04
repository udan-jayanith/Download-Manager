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
	console.assert(errObj.error != undefined, 'error is undefined')
	return chrome.notifications.create(null, {
		title: 'Error',
		message: errObj.error,
		iconUrl: 'http://localhost:1616/pages/favicon.png',
		type: 'basic',
	})
}

function hideEl(el) {
	el.classList.add('hide')
}

function showEl(el) {
	el.classList.remove('hide')
}

function toggleHide(el) {
	el.classList.toggle('hide')
}

function decimalPoints(number, decimalPoints) {
	console.assert(typeof number == 'number', "number must be a number of type 'number'.")
	let str = String(number).split('.', 2)
	if (str.length == 1 || decimalPoints <= 0) {
		return Number(str[0])
	}
	return Number(str[0] + '.' + str[1].split('').splice(0, decimalPoints).join(''))
}

function dateAndTimeAgo(dateAndTime) {
	let dt = luxon.DateTime.now().minus(luxon.DateTime.fromISO(dateAndTime).c).c
	if (dt.year > 0) {
		return `${dt.day}/${dt.month}/${dt.year} ago`
	} else if (dt.month > 0) {
		return `${dt.day}d ${dt.month}m ago`
	} else if (dt.day > 0) {
		return `${dt.day}d ago`
	} else if (dt.hour > 0) {
		return `${dt.hour}h ago`
	} else if (dt.minute > 0) {
		return `${dt.minute}m ago`
	}
	return `${dt.second}s ago`
}

function EventDelegation(parentElement, elementSelector, eventType, callback) {
	parentElement.addEventListener(eventType, (e) => {
		let el = e.target.closest(elementSelector)
		if (el == null) {
			return
		}
		callback(el, e)
	})
}
