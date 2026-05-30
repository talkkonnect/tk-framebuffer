package main

// signalLevel returns 0..1 for the single Signal bar (full scale on TX or RX).
func signalLevel(st DisplayState) float64 {
	if st.Transmitting || st.Receiving {
		return 1.0
	}
	return 0.0
}
