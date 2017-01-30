package audio

// Buffer de audio Jack
type AudioBuffer []float32
type AudioBuffers []AudioBuffer

/*
 * Retorna o número de samples de um canal
 */
func (b AudioBuffer) NumSamples() int {
	return len(b)
}

/*
 * Retorna um sample
 */
func (b AudioBuffer) Sample(i int) float32 {
	return b[i]
}

/*
 * Retorna o número de canais de um buffer de audio
 */
func (b AudioBuffers) NumChannels() int {
	return len(b)
}

/*
 * Retorna um canal de um buffer de audio
 */
func (b AudioBuffers) Channel(c int) AudioBuffer {
	return b[c]
}

/*
 * Retorna o número de samples de um buffer de audio
 */
func (b AudioBuffers) NumSamples() int {
	return b.Channel(0).NumSamples()
}

/*
 * Cria um canal de audio
 */
func NewAudioBuffer(numSamples int) AudioBuffer {
	return AudioBuffer(make([]float32, numSamples))
}

/*
 * Cria um buffer de audio
 */
func NewAudioBuffers(numChannels, numSamples int) AudioBuffers {
	audio := AudioBuffers(make([]AudioBuffer, numChannels))

	for i := 0; i < numChannels; i++ {
		audio[i] = NewAudioBuffer(numSamples)
	}

	return audio
}
