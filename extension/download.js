//https://developer.chrome.com/docs/extensions/reference/api/downloads#type-DownloadOptions
/*
chrome.downloads.download({
	filename: 'test/image.jpg',
	url: 'https://upload.wikimedia.org/wikipedia/commons/thumb/4/41/Sunflower_from_Silesia2.jpg/1200px-Sunflower_from_Silesia2.jpg',
})

*/

let download = {
    getNewDownloadId: async function () {
        let downloadIdObject = (await chrome.storage.local.get(['download-id']))
        if (Object.keys(downloadIdObject).length == 0) {
            await chrome.storage.local.set({
                downloadId: 1
            })
            return 0
        }
        let downloadId = downloadIdObject.downloadId
        await chrome.storage.local.set({
            downloadId: downloadId+1
        })
        return downloadId
    }, 
	download: function (obj, settings) {
        
    },
}

function sequentialDownload() {}
function parallelDownload() {}
