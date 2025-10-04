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
