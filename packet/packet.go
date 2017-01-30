package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/vtfr/gotrip/audio"
	"io"
)

type PacketType byte

const (
	HandshakePacketId PacketType = iota
	AudioPacketId
	ExitPacketId
)

//type PacketMetadata struct {
//	Type PacketType
//	Size uint32
//}

type HandshakePacket struct {
	NumChannels uint8
}

//type AudioPacket struct {
//	Audio AudioBuffers
//}

/*
 * Escreve um pacote de handshake
 *
 * Tipo de Pacote (1 byte)
 * Número de Canais (1 byte)
 */
func WriteHandshakePacket(numChannels uint8) *bytes.Buffer {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.LittleEndian, HandshakePacketId)
	binary.Write(buffer, binary.LittleEndian, numChannels)
	return buffer
}

/*
 * Lê um pacote de handshake
 */
func ReadHandshakePacket(reader io.Reader) (packet HandshakePacket, err error) {
	err = binary.Read(reader, binary.LittleEndian, &packet.NumChannels)
	return
}

/*
 * Escreve um pacote de audio
 *
 * Tipo de Pacote (1 byte)
 * Tamanho do pacote em bytes (4 bytes)
 *
 * Número de Canais (1 byte)
 * Número de samples (4 bytes)
 * Buffers de audio (um canal por vez) (Número de Canais
 *    * Número de samples * Sizeof(float32))
 */
func WriteAudioPacket(audio audio.AudioBuffers) *bytes.Buffer {
	buffer := new(bytes.Buffer)

	//	packetLen := audio.NumChannels() * audio.NumSamples() * 4 + 5;

	binary.Write(buffer, binary.LittleEndian, byte(AudioPacketId))
	//	binary.Write(buffer, binary.LittleEndian, uint32(packetLen))
	binary.Write(buffer, binary.LittleEndian, uint8(audio.NumChannels()))
	binary.Write(buffer, binary.LittleEndian, uint32(audio.NumSamples()))

	// Escreve faixa por faixa de audio
	for i := 0; i < audio.NumChannels(); i++ {
		binary.Write(buffer, binary.LittleEndian, []float32(audio.Channel(i)))
	}

	return buffer
}

/*
 * Lê um pacote de audio
 */
func ReadAudioPacket(reader io.Reader) (audio.AudioBuffers, error) {
	var numChannels uint8
	var numSamples uint32

	if err := binary.Read(reader, binary.LittleEndian, &numChannels); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &numSamples); err != nil {
		return nil, err
	}

	buffer := audio.NewAudioBuffers(int(numChannels), int(numSamples))
	for i := 0; i < buffer.NumChannels(); i++ {
		err := binary.Read(reader, binary.LittleEndian, []float32(buffer[i]))
		if err != nil {
			return nil, err
		}
	}

	return buffer, nil
}

/*
 * Lê o tipo de pacote
 */
func ReadPacketType(reader io.Reader) (t PacketType, err error) {
	err = binary.Read(reader, binary.LittleEndian, &t)
	return
}

///*
// * Escreve um pacote de audio
// */
//func NewAudioPacket(audio AudioBuffers) (*bytes.Buffer, error) {
//	buffer := new(bytes.Buffer);
//	buffer.WriteByte(AudioPacket); // Tipo de Pacote (1 byte)
//
//	err := binary.Write(buffer, binary.LittleEndian, audio);
//	if err != nil {
//		return nil, err
//	}
//
//	return buffer, nil
//}
