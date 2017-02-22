package audio

// Audio Buffer
type Buffer []float32

// Return the number of samples of this buffer
func (b Buffer) NumSamples() int {
	return len(b)
}

// Return a sample from this buffer
func (b Buffer) Sample(i int) float32 {
	return b[i]
}

// Audio Frame
type Frame []Buffer

// Return the number of channels of this frame
func (b Frame) NumChannels() int {
	return len(b)
}

// Return a channel from this frame
func (b Frame) Channel(c int) Buffer {
	return b[c]
}

// Return the number of samples of this frame
func (b Frame) NumSamples() int {
	return b.Channel(0).NumSamples()
}

// Creates a new audio buffer
func NewBuffer(samples int) Buffer {
	return Buffer(make([]float32, samples))
}

// Creates a new frame
func NewFrame(channels, samples int) Frame {
	audio := Frame(make([]Buffer, channels))

	for i := 0; i < channels; i++ {
		audio[i] = NewBuffer(samples)
	}

	return audio
}
