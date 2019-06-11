package main

type Encoder interface {
	getResult(count int) uint
	initGF(poly uint)
	initCode(nsym uint, index uint)
	encode(len uint, data []uint)
}

type ReedSolomon struct {
	logmod uint
	rlen   uint
	logt   []uint
	alog   []uint
	rspoly []uint
	res    []uint
}

func NewReedSolomon() *ReedSolomon {
	return &ReedSolomon{}
}
func (r *ReedSolomon) getResult(count int) uint {
	return r.res[count]
}

func (r *ReedSolomon) initGF(poly uint) {

	var b, p, m, v uint
	// Find the top bit, and hence the symbol size
	m = 0
	for b = 1; b <= poly; b <<= 1 {
		m++
	}
	b >>= 1
	m--

	// Calculate the log/alog tables
	r.logmod = (1 << m) - 1
	r.logt = make([]uint, r.logmod+1)
	r.alog = make([]uint, r.logmod)

	p = 1
	for v = 0; v < r.logmod; v++ {
		r.alog[v] = p
		r.logt[p] = v
		p <<= 1
		if (p & b) != 0 {
			p ^= poly
		}
	}
}

func (r *ReedSolomon) initCode(nsym uint, index uint) {
	var k, i uint

	r.rspoly = make([]uint, nsym+1)

	r.rlen = nsym

	r.rspoly[0] = 1
	for i = 1; i <= nsym; i++ {
		r.rspoly[i] = 1
		for k = i - 1; k > 0; k-- {
			if r.rspoly[k] != 0 {
				r.rspoly[k] = r.alog[(r.logt[r.rspoly[k]]+index)%r.logmod]
			}
			r.rspoly[k] ^= r.rspoly[k-1]
		}
		r.rspoly[0] = r.alog[(r.logt[r.rspoly[0]]+index)%r.logmod]
		index++
	}
}

func (r *ReedSolomon) encode(len uint, data []uint) {
	var i, k, m uint

	r.res = make([]uint, r.rlen)

	for i = 0; i < r.rlen; i++ {
		r.res[i] = 0
	}

	for i = 0; i < len; i++ {
		m = r.res[r.rlen-1] ^ data[i]
		for k = r.rlen - 1; k > 0; k-- {
			if (m != 0) && (r.rspoly[k] != 0) {
				r.res[k] = r.res[k-1] ^ r.alog[(r.logt[m]+r.logt[r.rspoly[k]])%r.logmod]
			} else {
				r.res[k] = r.res[k-1]
			}
		}

		if (m != 0) && (r.rspoly[0] != 0) {
			r.res[0] = r.alog[(r.logt[m]+r.logt[r.rspoly[0]])%r.logmod]
		} else {
			r.res[0] = 0
		}
	}
}
