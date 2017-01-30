package audio

import (
	jackc "github.com/xthexder/go-jack"
	"strconv"
)

var facade struct {
	SendChannel    chan AudioBuffers
	ReceiveChannel chan AudioBuffers

	clientName  string
	client      *jackc.Client
	inputPorts  []*jackc.Port
	outputPorts []*jackc.Port
}

// Retorna o número de portas de entrada do cliente global
func NumInputPorts() int {
	return len(facade.inputPorts)
}

// Retorna o número de portas de saída do cliente global
func NumOutputPorts() int {
	return len(facade.outputPorts)
}

// Retorna o nome do cliente global
func ClientName() string {
	return facade.clientName
}

// Canal de envio do cliente global
func SendChannel() chan AudioBuffers {
	return facade.SendChannel
}

// Canal de recebimento do cliente global
func ReceiveChannel() chan AudioBuffers {
	return facade.ReceiveChannel
}

// Inicializa o cliente
func Initialize(name string, numInputPorts, numOutputPorts int) error {
	// Define algumas informações globais e cria as portas e canais
	facade.clientName = name
	facade.inputPorts = make([]*jackc.Port, numInputPorts)
	facade.outputPorts = make([]*jackc.Port, numOutputPorts)
	facade.SendChannel = make(chan AudioBuffers, 2)
	facade.ReceiveChannel = make(chan AudioBuffers, 2)

	// Abre um novo cliente jack
	client, status := jackc.ClientOpen(name, jackc.NoStartServer)
	if status != 0 {
		return jackc.Strerror(status)
	}

	// Define o cliente
	facade.client = client

	// Define o callback
	if code := client.SetProcessCallback(processCallback); code != 0 {
		return jackc.Strerror(code)
	}

	// Tenta ativar o cliente Jack
	if code := client.Activate(); code != 0 {
		return jackc.Strerror(code)
	}

	// Configura as portas de entrada do cliente Jack
	for i := 0; i < numInputPorts; i++ {
		port := client.PortRegister("in_"+strconv.Itoa(i),
			jackc.DEFAULT_AUDIO_TYPE,
			jackc.PortIsInput, 0)

		facade.inputPorts[i] = port
	}

	// Configura as portas de saída do cliente Jack
	for i := 0; i < numOutputPorts; i++ {
		port := client.PortRegister("out_"+strconv.Itoa(i),
			jackc.DEFAULT_AUDIO_TYPE,
			jackc.PortIsOutput, 0)

		facade.outputPorts[i] = port
	}

	// Ativa o cliente
	if code := facade.client.Activate(); code != 0 {
		return jackc.Strerror(code)
	}

	// Inicia a gorotina de escrita
	go func() {
		for {
			buffer := <-facade.SendChannel
			nFrames := buffer.NumSamples()

			for portId, port := range facade.outputPorts {
				portBuffer := port.GetBuffer(uint32(nFrames))

				// Copia os buffers
				for i := 0; i < nFrames; i++ {
					portBuffer[i] = jackc.AudioSample(buffer[portId][i])
				}
			}
		}
	}()

	return nil
}

// Callback de processamento do cliente global
func processCallback(nFrames uint32) int {
	buffer := NewAudioBuffers(NumInputPorts(), int(nFrames))
	for portId, port := range facade.inputPorts {
		portBuffer := port.GetBuffer(nFrames)

		for i := uint32(0); i < nFrames; i++ {
			buffer[portId][i] = float32(portBuffer[i])
		}

		select {
		case facade.ReceiveChannel <- buffer:
		default:
		}
	}

	return 0
}
