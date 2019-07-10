package australiapost

// Encoder generic encoder interface
type Encoder interface {
	Encode(data []uint) []uint
}

type reedSolomon struct {
	log         []uint
	alog        []uint
	polynomials []uint
	polynomial  uint
	gf          uint
}

// NewReedSolomon create a new instance of Reed Solomon
func NewReedSolomon() Encoder {
	return &reedSolomon{
		log: []uint{
			0, 0, 1, 6, 2, 12, 7, 26,
			3, 32, 13, 35, 8, 48, 27, 18,
			4, 24, 33, 16, 14, 52, 36, 54,
			9, 45, 49, 38, 28, 41, 19, 56,
			5, 62, 25, 11, 34, 31, 17, 47,
			15, 23, 53, 51, 37, 44, 55, 40,
			10, 61, 46, 30, 50, 22, 39, 43,
			29, 60, 42, 21, 20, 59, 57, 58,
		},
		alog: []uint{
			1, 2, 4, 8, 16, 32, 3, 6,
			12, 24, 48, 35, 5, 10, 20, 40,
			19, 38, 15, 30, 60, 59, 53, 41,
			17, 34, 7, 14, 28, 56, 51, 37,
			9, 18, 36, 11, 22, 44, 27, 54,
			47, 29, 58, 55, 45, 25, 50, 39,
			13, 26, 52, 43, 21, 42, 23, 46,
			31, 62, 63, 61, 57, 49, 33,
		},
		polynomials: []uint{48, 17, 29, 30, 1},
		polynomial:  67,
		gf:          64,
	}
}

// Encode encode data using Reed Solomon
func (r *reedSolomon) Encode(data []uint) []uint {
	var i, k, m, l uint

	v := make([]uint, 4)

	for i = 0; i < 4; i++ {
		v[i] = 0
	}

	l = uint(len(data))

	for i = 0; i < l; i++ {
		m = v[3] ^ data[i]
		for k = 3; k > 0; k-- {
			if (m != 0) && (r.polynomials[k] != 0) {
				v[k] = v[k-1] ^ r.alog[(r.log[m]+r.log[r.polynomials[k]])%(r.gf-1)]
			} else {
				v[k] = v[k-1]
			}
		}

		if (m != 0) && (r.polynomials[0] != 0) {
			v[0] = r.alog[(r.log[m]+r.log[r.polynomials[0]])%(r.gf-1)]
		} else {
			v[0] = 0
		}
	}

	return []uint{v[3], v[2], v[1], v[0]}
}
