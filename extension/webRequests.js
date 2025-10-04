//This takes a header value and tokenizes it and returns it.
function parseHeaderValue(headerValue) {
	let tokens = []

	let currentToken = ''
	let isReadingQuote = false
	for (let i = 0; i < headerValue.length; i++) {
		if (!isReadingQuote && headerValue[i] == ' ') {
			continue
		}

		switch (headerValue[i]) {
			case `"`:
				isReadingQuote = !isReadingQuote
				break
			case ',':
			case `;`:
				tokens.push(currentToken)
				currentToken = ''
				break
			default:
				currentToken += headerValue[i]
				break
		}
	}
	tokens.push(currentToken)
	return tokens
}

function escapeFolderName(folderName) {
	let newFolderName = ''
	let invalidChar = new Set(['/', '\\'[0], '=', '"', ',', ':', '*', '+', '<', '>', '|', '.'])
	for (let i = 0; i < folderName.length; i++) {
		if (invalidChar.has(folderName[i])) {
			continue
		}
		newFolderName += folderName[i]
	}
	return newFolderName
}

//This takes the content type value (Ex: application/json) and returns the extensionName(json).
function mediaTypeExtensionName(mediaTypeHTTPHeaderValue) {
	let extensionName = ''
	if (mediaTypeHTTPHeaderValue == undefined) {
		return extensionName
	}
	for (let i = mediaTypeHTTPHeaderValue.length - 1; i >= 0; i--) {
		if (mediaTypeHTTPHeaderValue[i] == '/') {
			break
		}
		extensionName = mediaTypeHTTPHeaderValue[i] + extensionName
	}
	return extensionName.toLowerCase()
}

//parseURL takes a url and returns  extensionName and the fileName in a object.
function parseURL(url) {
	let obj = {
		fileName: '',
		extensionName: '',
	}

	url = new URL(url).pathname.trimEnd()

	for (let i = url.length - 1; i >= 0; i--) {
		if (url[i] == '.' && obj.extensionName == '') {
			obj.extensionName = obj.fileName
			obj.fileName = ''
			continue
		} else if (url[i] == '/') {
			break
		}
		obj.fileName = url[i] + obj.fileName
	}
	obj.fileName = escapeFolderName(obj.fileName)
	obj.extensionName = escapeFolderName(obj.extensionName)
	return obj
}

function getHeaderValue(headers, headerName) {
	for (let i = 0; i < headers.length; i++) {
		if (headers[i].name.toLowerCase() == headerName.toLowerCase()) {
			if (headers[i].value == null) {
				break
			}
			return headers[i].value
		}
	}
	return ''
}

function getContentTypeHeaderValue(headers) {
	return getHeaderValue(headers, 'Content-Type')
}

function randomString(length) {
	var result = ''
	var characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
	var charactersLength = characters.length
	for (var i = 0; i < length; i++) {
		result += characters.charAt(Math.floor(Math.random() * charactersLength))
	}
	return result
}

function transformDetails(details, headers) {
	let obj = {
		fileName: '',
		extensionName: '',
		url: details.url,
		method: details.method,
		statusCode: details.statusCode,
		type: details.type,
		tabId: details.tabId,
		contentLength: getHeaderValue(details.responseHeaders, 'Content-Length'),
		headers: headers,
		requestId: details.requestId,
	}

	let contentTypeHederValue = getContentTypeHeaderValue(details.responseHeaders)
	if (contentTypeHederValue == null) {
		return obj
	}
	let urlObj = parseURL(details.url)
	if (urlObj.extensionName == '') {
		urlObj.extensionName = mediaTypeExtensionName(parseHeaderValue(contentTypeHederValue)[0])
	}
	if (urlObj.fileName == '') {
		urlObj.fileName = randomString(5)
	}
	obj.extensionName = urlObj.extensionName
	obj.fileName = urlObj.fileName

	return obj
}

let webRequests = {
	arrayCap: 256,
	webRequestsContainer: new Map(),
	requestHeaders: new Map(),
	add: function (tabId, obj) {
		let array = this.webRequestsContainer.get(tabId)
		if (array == undefined) {
			array = []
			this.webRequestsContainer.set(tabId, array)
		}
		array.unshift(obj)
		if (array.length >= this.arrayCap) {
			array.splice(this.arrayCap, array.length - this.arrayCap)
		}
	},
	get: function (tabId) {
		let array = this.webRequestsContainer.get(tabId)
		if (array == undefined) {
			array = []
		}
		return array
	},
	delete: function (tabId) {
		this.webRequestsContainer.delete(tabId)
	},
}

chrome.webRequest.onBeforeSendHeaders.addListener(
	async ({requestId, requestHeaders}) => {
		webRequests.requestHeaders.set(requestId, requestHeaders)
	},
	{urls: ['<all_urls>']},
	['requestHeaders']
)

chrome.webRequest.onHeadersReceived.addListener(
	async (details) => {
		if (details.tabId == -1) {
			return
		}
		let headers = webRequests.requestHeaders.get(details.requestId)
		if (headers == undefined) {
			headers = []
		}
		let obj = transformDetails(details, headers)
		if (obj.extensionName == '' && obj.fileName == '') {
			return
		}
		let settings = await getSettings()
		if (settings.logWebRequest) {
			console.log(obj)
		}
		webRequests.add(details.tabId, obj)
	},
	{urls: ['<all_urls>']},
	['responseHeaders']
)

chrome.webRequest.onCompleted.addListener(
	async ({requestId}) => {
		webRequests.requestHeaders.delete(requestId)
	},
	{urls: ['<all_urls>']}
)

chrome.tabs.onUpdated.addListener((tabId, changeInfo) => {
	if (changeInfo.url == undefined) {
		return
	}
	webRequests.delete(tabId)
})

let currentTabId = -1

chrome.tabs.onActivated.addListener((activeInfo) => {
	currentTabId = activeInfo.tabId
})

message.onRequest('webRequests', (_, response) => {
	response({
		webRequest: webRequests.get(currentTabId),
	})
})
