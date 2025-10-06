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
	//2025-10-05T13:01:37+05:30
	dateAndTime = luxon.DateTime.fromISO(dateAndTime).c
	let diff =
		luxon.Duration.fromObject(luxon.DateTime.now().c).as('seconds') -
		luxon.Duration.fromObject(dateAndTime).as('seconds')

	let table = {
		seconds: Math.round(diff),
		minutes: function () {
			let res = Math.round(this.seconds / 60)
			return JSON.stringify(res) == 'null' ? 0 : res
		},
		hours: function () {
			let res = Math.round(this.minutes() / 60)
			return JSON.stringify(res) == 'null' ? 0 : res
		},
		days: function () {
			let res = Math.round(this.hours() / 24)
			return JSON.stringify(res) == 'null' ? 0 : res
		},
		months: function () {
			let res = Math.round(this.days() / 30)
			return JSON.stringify(res) == 'null' ? 0 : res
		},
		years: function () {
			let res = Math.round(this.months() / 12)
			return JSON.stringify(res) == 'null' ? 0 : res
		},
	}

	if (table.minutes() <= 0) {
		return `${table.seconds}s ago`
	} else if (table.hours() <= 0) {
		return `${table.minutes()}m ${
			table.seconds - table.minutes() * 60 > 0 ? table.seconds - table.minutes() * 60 + 's ' : ''
		}ago`
	} else if (table.days() <= 0) {
		return `${table.hours()}h ${
			table.minutes() - table.hours() * 60 > 0 ? table.minutes() - table.hours() * 60 + 'm ' : ''
		}ago`
	} else if (table.months() <= 0) {
		return `${table.days()}d ${
			table.hours() - table.days() * 24 > 0 ? table.hours() - table.days() * 24 + 'h ' : ''
		}ago`
	} else if (table.years() <= 0) {
		return `${table.months()}m ${
			table.days() - table.months() * 30 > 0 ? table.days() - table.months() * 30 + 'd ' : ''
		}ago`
	}

	return `${dateAndTime.day}-${dateAndTime.month}-${dateAndTime.year}`
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

componentSystem.rootDir = './components/'

function DeleteElementWithAnimation(element) {
	console.assert(element != null, 'Element is null')
	console.assert(typeof element == 'object', 'element is not a HTML DOM element')

	let elementCopy = element.cloneNode(true)
	elementCopy.classList.add('delete-animation')

	element.parentNode.replaceChild(elementCopy, element)
	element.remove()

	elementCopy.addEventListener('animationend', ({target}) => {
		target.remove()
	})
}
