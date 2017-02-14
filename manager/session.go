package manager

import (
	"bytes"
	"github.com/vtfr/gotrip/audio"
	"github.com/vtfr/gotrip/packet"
	"log"
	"net"
	"fmt"
)

type Mode uint8

const (
	Master Mode = iota
	Member
)

type Session struct {
	Mode Mode
	Name string

	// Informações do streamming de audio
	NumChannels uint32
	SampleRate  uint64

	// Streammer de audio
	AudioClient *audio.Client

	// Informações da conexão
	LocalAddr   *net.UDPAddr
	RemoteAddr  *net.UDPAddr
	Listener    *net.UDPConn
	Connection  *net.UDPConn

	// Flag de inicio de streamming ou não
	Streamming  bool

	// Canal para termino de sessão
	Exit chan bool
}

func NewSessionAsMaster(name, addr string, numChan, sampleRate int) (*Session, error) {
	s := &Session {
		Mode: Master,
		Name: name,
		NumChannels: uint32(numChan),
		SampleRate: uint64(sampleRate),
		Streamming: false,
		Exit: make(chan bool),
	}

	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}

	s.LocalAddr = laddr
	s.Listener = conn
	return s, nil
}

func NewSessionAsMember(name, addr string) (*Session, error) {
	s := &Session{
		Mode: Member,
		Name: name,
		Streamming: false,
		Exit: make(chan bool),
	}

	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	laddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", raddr.Port + 1))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", s.LocalAddr)
	if err != nil {
		return nil, err
	}

	s.Listener = conn
	s.RemoteAddr = raddr
	s.LocalAddr = laddr
	return s, nil
}

func (s Session) String() string {
	if s.Streamming {
		return fmt.Sprintf("Session %s <-> ...", s.LocalAddr)
	}

	return fmt.Sprintf("%s <-> %s", s.LocalAddr, s.RemoteAddr)
}

func (s *Session) Run() {
	// Espera o handshake
	s.waitForHandshake()

	// Cria um cliente de audio com as configurações requisitadas
	log.Println("Creating audio interface...")
	c, err := audio.CreateClient(s.Name, int(s.NumChannels))
	if err != nil {
		log.Fatalln("Failer creating audio client:", err)
	}

	s.AudioClient = c;

	// Inicia processamento
	go s.handleIncomming()
	go s.handleSending()
	<- s.Exit
}

func (s *Session) waitForHandshake() {
	// Envia o handshake inicial
	if s.Mode == Member {
		log.Println("Sending handshake...")
		s.Listener.
			WriteToUDP(packet.WriteHandshakeClient(packet.HandshakeClient{}).Bytes(),
			s.RemoteAddr)
	}

	log.Println("Waiting for handshake...");
	var buffer [256]byte
	for {
		// Escuta um pacote UDP
		log.Println("Receiving packet...")
		n, addr, err := s.Listener.ReadFromUDP(buffer[:])
		if err != nil {
			log.Println("Failed receiving packet from", addr, ":", err)
			continue;
		}

		// Inicia processamento do handshake
		log.Println("Reading packet...")
		p, err := packet.ReadPacket(bytes.NewBuffer(buffer[:n]))
		if err != nil {
			log.Println("Failed processing packet from", addr, ":", err)
			continue;
		}

		// Processa se o pacote é o handshake
		log.Println("Processing packet...")
		switch s.Mode {
		case Master:
			// Verifica se o pacote é um handshake client
			if _, ok := p.(packet.HandshakeClient); ok {
				log.Println("Handshake from client received!")
				// Envia informações do handshake
				s.RemoteAddr = addr;
				s.Streamming = true;

				s.Listener.WriteToUDP(packet.WriteHandshakeServer(packet.HandshakeServer{
						NumChannels: uint8(s.NumChannels),
						SampleRate: uint32(s.SampleRate),
					}).Bytes(), s.RemoteAddr)
				return;
			}

			log.Println("Invalid packet from", addr)
			continue;
		case Member:
			// Verifica se o pacote é um handshake client
			if hss, ok := p.(packet.HandshakeServer); ok {
				log.Println("Handshake from server received!")
				// Recee as informações do handshake
				s.NumChannels = uint32(hss.NumChannels);
				s.SampleRate  = uint64(hss.SampleRate);
				s.Streamming  = true;
				return
			}

			log.Println("Invalid packet from", addr)
			continue;
		}
	}
}

func (s *Session) handleIncomming() {
	log.Println("Processing incomming...")

	var buffer [16384]byte

	for {
		// Escuta um pacote UDP
		n, addr, err := s.Listener.ReadFromUDP(buffer[:])
		if err != nil {
			log.Println("Falha lendo pacote:", err)
			continue
		}

		// Verifica se esse pacote é dessa sessão
//		if !addr.IP.Equal(s.RemoteAddr.IP) {
//			log.Println("Ignoring packet from", addr)
//			continue;
//		}

		// Processa pacote
		p, err := packet.ReadPacket(bytes.NewBuffer(buffer[:n]))
		if err != nil {
			log.Println("Failed processing packet from", addr, ":", err)
			continue;
		}

		// Processa o pacote
		switch p.(type) {
		case packet.AudioFrame:
			frame := p.(packet.AudioFrame)

			select {
			case s.AudioClient.Send <- frame.Frame:
			default:
			}
		default:
			log.Println("Invalid packet", p, ". Ignoring")
		}
	}
}

func (s *Session) handleSending() {
	for {
		frame := <- s.AudioClient.Receive

		s.Listener.WriteToUDP(packet.WriteAudioFrame(packet.AudioFrame{
			Frame: frame,
		}).Bytes(), s.RemoteAddr)
	}
}
