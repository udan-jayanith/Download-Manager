let componentSystem = {
	rootDir: './',
	loadComponent: async function (relativePath, rootDir = this.rootDir) {
		let res = await fetch(rootDir + relativePath)
		let body = res.body.getReader()
		let content = new Array()
		while (true) {
			let {value, done} = await body.read()
			if (done) {
				break
			}
			content.push(...value)
		}
		let blob = new Blob([new Uint8Array(content)], {
			type: 'text/html',
		})
		let text = await blob.text()
		let templateDocument = Document.parseHTMLUnsafe('')
        templateDocument.body.innerHTML = text
        
        let templateEl = templateDocument.querySelector('body template')
		console.assert(templateEl != null, 'template is missing')
        return templateEl.content.cloneNode(true)
	},
}
