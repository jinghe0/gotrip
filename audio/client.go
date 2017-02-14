package audio

import (
	"github.com/xthexder/go-jack"
	"fmt"
)

var client *jack.Client
var clients []*Client

type Client struct {
	Name string
	NumChannels uint32
	SampleRate uint64
	Input  []*jack.Port
	Output []*jack.Port
	Receive chan Frame
	Send    chan Frame
}

func (c Client) NumInput() int {
	return len(c.Input)
}

func (c Client) NumOutput() int {
	return len(c.Output)
}

// Inicializa
func Initialize(name string) error {
	// Abre um novo cliente jack
	c, status := jack.ClientOpen(name, jack.NoStartServer)
	if status != 0 {
		return jack.Strerror(status)
	}

	// Define o cliente
	client = c

	// Define o callback
	if code := client.SetProcessCallback(processCallback); code != 0 {
		return jack.Strerror(code)
	}

	// Tenta ativar o cliente Jack
//	if code := client.Activate(); code != 0 {
//		return jack.Strerror(code)
//	}

	go processSending()
	return nil
}

// Cria novo cliente
func CreateClient(name string, numChannels int) (*Client, error) {
	c := &Client{
		Name: name,
		NumChannels: uint32(numChannels),
	}

	c.Input  = make([]*jack.Port, numChannels)
	c.Output = make([]*jack.Port, numChannels)
	c.Send    = make(chan Frame, 2)
	c.Receive = make(chan Frame, 2)

	// Cria as portas
	for i := 0; i < numChannels; i++ {
		in := client.PortRegister(fmt.Sprintf("%s-in-%d", name, i),
			jack.DEFAULT_AUDIO_TYPE,
			jack.PortIsInput, 0)
		out := client.PortRegister(fmt.Sprintf("%s-out-%d", name, i),
			jack.DEFAULT_AUDIO_TYPE,
			jack.PortIsOutput, 0)

		c.Input[i]  = in
		c.Output[i] = out
	}

	clients = append(clients, c)

	client.Activate()

	return c, nil
}

// Processa envio
func processSending() {
	for {
		for _, c := range clients {
			select {
			// Tenta ler o buffer sem travar
			case f := <- c.Send:
				// Para cada porta de saída
				for id, port := range c.Output {
					// Pega o buffer do jack
					b := port.GetBuffer(uint32(f.NumSamples()))
					// Copia os samples manualmente
					for i := 0; i < int(f.NumSamples()); i++ {
						b[i] = jack.AudioSample(f[id][i])
					}
				}
			default:
			}
		}
	}
}

// Callback de processamento do cliente global
func processCallback(nFrames uint32) int {
	for _, c := range clients {
		frame := NewFrame(int(c.NumChannels), int(nFrames))
		for i, port := range c.Input {
			buf := port.GetBuffer(nFrames)
			for k := 0; k < int(nFrames); k++ {
				frame[i][k] = float32(buf[k])
			}
		}

		select {
		case c.Receive <- frame:
		default:
		}
	}

	return 0
}

/*
import (
	"github.com/gordonklaus/portaudio"
)

func Initialize() {
	audio.Initialize()
}

func Terminate() {
	audio.Terminate()
}

// Audio Client
type Client struct {
	*portaudio.Stream
	Send    chan Frame
	Receive chan Frame
}

func (c Clinet) processAudio(in, out [][]float32) {
	// Verifica se audio precisa ser enviado
	select {
	case frame := <- c.Send
		for c := range out {
			for i := range out[c] {
				out[c][i] = frame[c][i]
			}
		}
	default:
	}
}





func NewClinet



func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()
	e := newEcho(time.Second / 3)
	i := newEcho(time.Second / 3)
	o := newEcho(time.Second / 3)
	defer e.Close()
	chk(e.Start())
	chk(i.Start())
	chk(o.Start())
	time.Sleep(120 * time.Second)
	chk(e.Stop())
}

type echo struct {
	*portaudio.Stream
	buffer []float32
	i      int
}

func newEcho(delay time.Duration) *echo {
	h, err := portaudio.HostApi(portaudio.JACK)
	chk(err)
	p := portaudio.LowLatencyParameters(h.DefaultInputDevice, h.DefaultOutputDevice)
	p.Input.Channels = 1
	p.Output.Channels = 1
	e := &echo{buffer: make([]float32, int(p.SampleRate*delay.Seconds()))}
	e.Stream, err = portaudio.OpenStream(p, e.processAudio)
	chk(err)
	return e
}

func (e *echo) processAudio(in, out []float32) {
	for i := range out {
		out[i] = .7 * e.buffer[e.i]
		e.buffer[e.i] = in[i]
		e.i = (e.i + 1) % len(e.buffer)
	}
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
*/









/*

import (
	jack "github.com/xthexder/go-jack"
	"strconv"
)

var facade struct {
	SendChannel    chan Frame
	ReceiveChannel chan Frame

	clientName  string
	client      *jack.Client
	Input  []*jack.Port
	Output []*jack.Port
}

// Retorna o número de portas de entrada do cliente global
func NumInput() int {
	return len(facade.Input)
}

// Retorna o número de portas de saída do cliente global
func NumOutput() int {
	return len(facade.Output)
}

// Retorna o nome do cliente global
func ClientName() string {
	return facade.clientName
}

// Canal de envio do cliente global
func SendChannel() chan Frame {
	return facade.SendChannel
}

// Canal de recebimento do cliente global
func ReceiveChannel() chan Frame {
	return facade.ReceiveChannel
}

// Envia audio para o cliente
func Send(a Frame) {
	facade.SendChannel <- a
}

// Envia audio para o cliente sem bloquear, dropando o buffer se necessário
func SendNonBlocking(a Frame) bool {
	select {
	case facade.SendChannel <- a:
		return true
	default:
		return false
	}
}

// Recebe o audio do cliente
func Receive(a Frame) Frame {
	return <-facade.ReceiveChannel
}

// Recebe audio do cliente sem bloquear, dropando o buffer se necessário
func ReceiveNonBlocking() Frame {
	select {
	case a := <- facade.ReceiveChannel:
		return a
	default:
		return nil
	}
}

// Inicializa o cliente
func Initialize(name string, numInput, numOutput int) error {
	// Define algumas informações globais e cria as portas e canais
	facade.clientName = name
	facade.Input = make([]*jack.Port, numInput)
	facade.Output = make([]*jack.Port, numOutput)
	facade.SendChannel = make(chan Frame, 2)
	facade.ReceiveChannel = make(chan Frame, 2)

	// Abre um novo cliente jack
	client, status := jack.ClientOpen(name, jack.NoStartServer)
	if status != 0 {
		return jack.Strerror(status)
	}

	// Define o cliente
	facade.client = client

	// Define o callback
	if code := client.SetProcessCallback(processCallback); code != 0 {
		return jack.Strerror(code)
	}

	// Tenta ativar o cliente Jack
	if code := client.Activate(); code != 0 {
		return jack.Strerror(code)
	}

	// Configura as portas de entrada do cliente Jack
	for i := 0; i < numInput; i++ {
		port := client.PortRegister("in_"+strconv.Itoa(i),
			jack.DEFAULT_AUDIO_TYPE,
			jack.PortIsInput, 0)

		facade.Input[i] = port
	}

	// Configura as portas de saída do cliente Jack
	for i := 0; i < numOutput; i++ {
		port := client.PortRegister("out_"+strconv.Itoa(i),
			jack.DEFAULT_AUDIO_TYPE,
			jack.PortIsOutput, 0)

		facade.Output[i] = port
	}

	// Ativa o cliente
	if code := facade.client.Activate(); code != 0 {
		return jack.Strerror(code)
	}

	// Inicia a gorotina de escrita
	go func() {
		for {
			buffer := <-facade.SendChannel
			nFrames := buffer.NumSamples()

			for portId, port := range facade.Output {
				portBuffer := port.GetBuffer(uint32(nFrames))

				// Copia os buffers
				for i := 0; i < nFrames; i++ {
					portBuffer[i] = jack.AudioSample(buffer[portId][i])
				}
			}
		}
	}()

	return nil
}

// Callback de processamento do cliente global
func processCallback(nFrames uint32) int {
	buffer := NewFrame(NumInput(), int(nFrames))
	for portId, port := range facade.Input {
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
*/
