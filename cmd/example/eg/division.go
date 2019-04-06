package eg

type Division struct {
	n float64
	d float64
}

func (d *Division) SetNumerator(n float64) {
	d.n = n
}

func (d *Division) SetDenominator(n float64) {
	d.d = n
}

func (d *Division) Quotient() float64 {
	return d.n / d.d
}