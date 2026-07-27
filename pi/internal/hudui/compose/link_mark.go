package compose

// LinkMarkFragment is the BLE link glyph (16×12 logical units, drawn at chrome link slot).
func LinkMarkFragment(connected bool) string {
	if connected {
		return `<path d="M2,9 L6,2 L6,6 L10,6 L6,10 L6,6" fill="none" stroke="#000" stroke-width="1.6" stroke-linejoin="miter"/>` +
			`<rect x="12" y="4" width="4" height="4" fill="#000"/>`
	}
	return `<line x1="2" y1="2" x2="10" y2="10" stroke="#000" stroke-width="1.6"/>` +
		`<line x1="10" y1="2" x2="2" y2="10" stroke="#000" stroke-width="1.6"/>`
}
