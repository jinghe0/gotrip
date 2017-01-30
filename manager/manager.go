package manager

import (
	"bytes"
	"github.com/vtfr/gotrip/audio"
	"github.com/vtfr/gotrip/packet"
	"log"
	"net"
)

type ConnectionMode byte

const (
	ManagerServerMode ConnectionMode = iota
	ManagerClientMode
)

type Manager struct {
	Mode           ConnectionMode
	StringAddress  string
	JackClientName string
}

/*
 * Cria um novo Manager
 */
func NewManager(jackClientName string, stringAddress string, isClient bool) (m Manager) {
	m.StringAddress = stringAddress
	m.JackClientName = jackClientName

	if isClient {
		m.Mode = ManagerClientMode
	} else {
		m.Mode = ManagerServerMode
	}

	err := audio.Initialize(jackClientName, 2, 2)
	fatalErr(err)

	return
}

/*
 * Inicia dinamicamente o manager de acordo com a configuração fornecida
 */
func (m Manager) Start() {
	switch m.Mode {
	case ManagerServerMode:
		m.StartServer()
	case ManagerClientMode:
		m.StartClient()
	}
}

/*
 * Em caso de erro, reporta erro fatal
 */
func fatalErr(err error) {
	if err != nil {
		log.Fatalln("Fatal error:", err)
	}
}

/*
 * Inicia o manager como servidor
 */
func (m Manager) StartServer() {
	//	receiveChannel := audio.ReceiveChannel()
	sendChannel := audio.SendChannel()

	// Resolve o endereço UDP fornecido e checa sua validade
	udpAddr, err := net.ResolveUDPAddr("udp", m.StringAddress)
	fatalErr(err)

	// Inicia um servidor UDP
	udpConn, err := net.ListenUDP("udp", udpAddr)
	fatalErr(err)

	defer udpConn.Close()

	// Buffer para leitura de pacotes
	buffer := make([]byte, 16392)

	// Espera a primera conexão ser feita, para realizar o Handshake.
	// Depois disso toda conexão que for feita com um IP diferente e uma sessão
	// diferente serão dropadas automaticamente.
	// var clientSession packet.Session;
	var clientIpAddr net.IP

	// Recebe a primeira conexão
	log.Println("Esperando handshake...")

	for {
		n, addr, err := udpConn.ReadFromUDP(buffer)
		fatalErr(err)

		buf := bytes.NewBuffer(buffer[:n])
		t, err := packet.ReadPacketType(buf)
		fatalErr(err)

		if t == packet.HandshakePacketId {
			log.Println("Handshake recebido")

			_, err := packet.ReadHandshakePacket(buf)
			fatalErr(err)

			// Define esse cliente como o cliente com o qual nos comunicaremos
			clientIpAddr = addr.IP
			break
		}
	}

	// Inicia troca de buffers
	for {
		n, addr, err := udpConn.ReadFromUDP(buffer)
		fatalErr(err)

		// Não processa se não vier do IP que originou o handshake
		if !clientIpAddr.Equal(addr.IP) {
			continue
		}

		// Lê o pacote
		buf := bytes.NewBuffer(buffer[:n])
		t, err := packet.ReadPacketType(buf)
		fatalErr(err)

		// Processa o pacote
		switch t {
		case packet.AudioPacketId:
			audio, err := packet.ReadAudioPacket(buf)
			fatalErr(err)

			log.Println("Pacote de audio recebido")

			select {
			case sendChannel <- audio:
			default:
			}
		}
	}
}

/*
 * Inicia o servidor no modo cliente
 */
func (m Manager) StartClient() {
	receiveChannel := audio.ReceiveChannel()
	//	sendChannel    := audio.SendChannel()

	// Tenta realizar o handshake
	log.Println("Esperando conexão...")
	conn, err := net.Dial("udp", m.StringAddress)
	fatalErr(err)

	defer conn.Close()

	log.Println("Enviando handshake...")
	_, err = packet.WriteHandshakePacket(2).WriteTo(conn)
	fatalErr(err)

	log.Println("Recebendo audio...")

	// GoRotina que recebe audio
	/*	go func() {
			for {
				conn, err := net.Dial("udp", address)
				fatalErr(err)

				defer conn.Close()
				audio, err = packet.ReadAudioPacket(conn);
				fatalErr(err)

				log.Println("Audio:")
				log.Println(audio.NumChannels())
				log.Println(audio.NumSamples())

				sendChannel <- audio

				time.Sleep(100 * time.Millisecond)
			}
		}
	*/

	// Rotina que envia audio
	for {
		audio := <-receiveChannel
		fatalErr(err)

		log.Println("Enviando pacote de audio")
		_, err = packet.WriteAudioPacket(audio).WriteTo(conn)
		fatalErr(err)
	}
}
