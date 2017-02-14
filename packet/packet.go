package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/vtfr/gotrip/audio"
	"io"
)

// Packet Type
type PacketType byte

// Types of packet
const (
	HandshakeClientId PacketType = iota
	HandshakeServerId
	AudioFrameId
)

// Packet send or received through the network
type Packet interface {}

// Handshake Client Packet
type HandshakeClient struct {
	Version uint8
}

// Handshake Server Packet
type HandshakeServer struct {
	Version     uint8
	Session     uint32
	NumChannels uint8
	SampleRate  uint32
}

// Audio Frame Packet
type AudioFrame struct {
	Id    uint8
	Frame audio.Frame
}

func WriteHandshakeClient(p HandshakeClient) *bytes.Buffer {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.LittleEndian, HandshakeClientId)
	binary.Write(buffer, binary.LittleEndian, p.Version)
	return buffer
}

func WriteHandshakeServer(p HandshakeServer) *bytes.Buffer {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.LittleEndian, HandshakeServerId)
	binary.Write(buffer, binary.LittleEndian, p.Version)
	binary.Write(buffer, binary.LittleEndian, p.Session)
	binary.Write(buffer, binary.LittleEndian, p.NumChannels)
	binary.Write(buffer, binary.LittleEndian, p.SampleRate)
	return buffer
}

func WriteAudioFrame(p AudioFrame) *bytes.Buffer {
	buffer := new(bytes.Buffer)

	binary.Write(buffer, binary.LittleEndian, AudioFrameId)
//	binary.Write(buffer, binary.LittleEndian, p.Id)
	binary.Write(buffer, binary.LittleEndian, uint32(p.Frame.NumChannels()))
	binary.Write(buffer, binary.LittleEndian, uint32(p.Frame.NumSamples()))

	// Escreve faixa por faixa de audio
	for i := 0; i < p.Frame.NumChannels(); i++ {
		binary.Write(buffer, binary.LittleEndian, []float32(p.Frame.Channel(i)))
	}

	return buffer
}

func ReadPacket(reader io.Reader) (Packet, error) {
	t, err := readPacketType(reader)
	if err != nil {
		return nil, err
	}

	switch t {
	case HandshakeClientId:
		return readHandshakeClient(reader)
	case HandshakeServerId:
		return readHandshakeServer(reader)
	case AudioFrameId:
		return readAudioFrame(reader)
	}

	return nil, nil
}

func readPacketType(reader io.Reader) (t PacketType, err error) {
	err = binary.Read(reader, binary.LittleEndian, &t)
	return
}

// reads a handshake client packet from the stream
func readHandshakeClient(reader io.Reader) (packet HandshakeClient, err error) {
	err = binary.Read(reader, binary.LittleEndian, &packet.Version)
	return
}

func readHandshakeServer(reader io.Reader) (packet HandshakeServer, err error) {
	err = binary.Read(reader, binary.LittleEndian, &packet.Version)
	if err != nil {
		return
	}

	err = binary.Read(reader, binary.LittleEndian, &packet.Session)
	if err != nil {
		return
	}

	err = binary.Read(reader, binary.LittleEndian, &packet.NumChannels)
	if err != nil {
		return
	}

	err = binary.Read(reader, binary.LittleEndian, &packet.SampleRate)
	if err != nil {
		return
	}

	return
}


// reads an audio frame packet from the stream
func readAudioFrame(reader io.Reader) (packet AudioFrame, err error) {
	var channels uint32
	var samples uint32

//	err = binary.Read(reader, binary.LittleEndian, &packet.Id)
//	if err != nil {
//		return
//	}

	err = binary.Read(reader, binary.LittleEndian, &channels)
	if err != nil {
		return
	}

	err = binary.Read(reader, binary.LittleEndian, &samples)
	if err != nil {
		return
	}

	// Read all audio channels
	frame := audio.NewFrame(int(channels), int(samples))
	for i := 0; i < frame.NumChannels(); i++ {
		err = binary.Read(reader, binary.LittleEndian, []float32(frame[i]))
		if err != nil {
			return
		}
	}

	packet.Frame = frame
	return
}
