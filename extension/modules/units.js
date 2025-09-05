function byte(byteAmount) {
	let obj = {
		byte: byteAmount,
		kb: function () {
			return this.byte / 1024
		},
		mb: function () {
			return this.kb() / 1024
		},
		gb: function () {
			return this.mb() / 1024
		},
		get: function () {
			let obj = {
				data: this.byte,
				unit: 'byte',
			}
			if (obj.data < 1024) {
				return obj
			} else if (this.kb() < 1024) {
				obj.data = this.kb()
				obj.unit = 'kb'
			} else if (this.mb() < 1024) {
				obj.data = this.mb()
				obj.unit = 'mb'
            } else {
                obj.data = this.gb()
                obj.unit = 'gb'
            }
            return obj
		},
	}
	return obj
}

//DateTime.fromISO(DateTime.now().toString()).toISODate()