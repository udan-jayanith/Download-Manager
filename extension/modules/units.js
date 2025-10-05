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
				unit: 'Byte',
			}
			if (obj.data < 1024) {
				return obj
			} else if (this.kb() < 1024) {
				obj.data = this.kb()
				obj.unit = 'KB'
			} else if (this.mb() < 1024) {
				obj.data = this.mb()
				obj.unit = 'MB'
            } else {
                obj.data = this.gb()
                obj.unit = 'GB'
            }
            return obj
		},
	}
	return obj
}
