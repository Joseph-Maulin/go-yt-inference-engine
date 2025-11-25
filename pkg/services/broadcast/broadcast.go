package broadcast

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Joseph-Maulin/go-yt-inference-engine/pkg/services/stream"
	"gocv.io/x/gocv"
)

const (
	startingBroadcastPort   = 9000
	maximumBroadcastStreams = 10
)

type Broadcast struct {
	YoutubeURL string
	Port       int
	frameChan  <-chan *gocv.Mat
	conn       *net.UDPConn
	cancel     context.CancelFunc
	ctx        context.Context
}

type BroadcastService struct {
	StreamService *stream.StreamService

	broadcasts map[string]*Broadcast
	mu         sync.RWMutex
}

func (b *BroadcastService) getNextPort() (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.broadcasts) >= maximumBroadcastStreams+1 {
		return 0, fmt.Errorf("maximum number of broadcasts reached")
	}

	// Find available port in range starting from port
	portFound := false
	var port int
	for port = startingBroadcastPort; port < startingBroadcastPort+maximumBroadcastStreams; port++ {
		conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
		if err != nil {
			continue
		}
		conn.Close()
		portFound = true
		break
	}
	if !portFound {
		return 0, fmt.Errorf("failed to find a free port")
	}
	return port, nil
}

func NewBroadcastService(streamService *stream.StreamService) *BroadcastService {
	return &BroadcastService{
		StreamService: streamService,
		broadcasts:    make(map[string]*Broadcast),
		mu:            sync.RWMutex{},
	}
}

func (b *BroadcastService) broadcastLoop(broadcast *Broadcast) error {
	defer broadcast.conn.Close()

	clients := make(map[string]*net.UDPAddr)
	var clientMutex sync.RWMutex

	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-broadcast.ctx.Done():
				log.Printf("Broadcast context done, stopping client listener; youtubeURL=%s\n", broadcast.YoutubeURL)
				return
			default:
				n, addr, err := broadcast.conn.ReadFromUDP(buf)
				if err != nil {
					broadcast.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

					if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
						continue
					}
					log.Printf("Error reading from UDP for client listener; youtubeURL=%s, error=%v", broadcast.YoutubeURL, err)
					continue
				}
				broadcast.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				msg_rcvd := string(buf[:n])
				switch msg_rcvd {
				case "CONNECT":
					clientMutex.Lock()
					clients[addr.String()] = addr
					clientMutex.Unlock()
					log.Printf("Client connected; youtubeURL=%s, client=%s", broadcast.YoutubeURL, addr.String())
				case "DISCONNECT":
					clientMutex.Lock()
					delete(clients, addr.String())
					clientMutex.Unlock()
					log.Printf("Client disconnected; youtubeURL=%s, client=%s", broadcast.YoutubeURL, addr.String())
				default:
					log.Printf("Unknown message received from client listener; youtubeURL=%s, message=%s", broadcast.YoutubeURL, msg_rcvd)
				}
			}
		}
	}()

	for {
		select {
		case <-broadcast.ctx.Done():
			log.Printf("Broadcast context done, stopping broadcast loop; youtubeURL=%s\n", broadcast.YoutubeURL)
			return nil
		case frame, ok := <-broadcast.frameChan:
			if !ok {
				log.Printf("Frame channel closed, stopping broadcast loop; youtubeURL=%s\n", broadcast.YoutubeURL)
				return nil
			}

			// Encode frame to JPEG
			frameBuf, err := gocv.IMEncode(".jpg", *frame)
			if err != nil {
				log.Printf("Error encoding frame to JPEG; youtubeURL=%s, error=%v", broadcast.YoutubeURL, err)
				frame.Close()
				continue
			}

			// Send frame to all clients
			clientMutex.RLock()
			for addrStr, addr := range clients {
				sendBufferSize := 65000
				data := frameBuf.GetBytes()

				if len(data) > sendBufferSize {
					log.Printf("Frame data too large to send to client; youtubeURL=%s, client=%s, dataSize=%d", broadcast.YoutubeURL, addrStr, len(data))
					continue
				}

				_, err := broadcast.conn.WriteToUDP(data, addr)
				if err != nil {
					log.Printf("Error sending frame to client; youtubeURL=%s, client=%s, error=%v", broadcast.YoutubeURL, addrStr, err)
					continue
				}
				log.Printf("Frame sent to client; youtubeURL=%s, client=%s, dataSize=%d", broadcast.YoutubeURL, addrStr, len(data))
			}
			clientMutex.RUnlock()

			frameBuf.Close()
			frame.Close()
		}
	}
}

func (b *BroadcastService) StartBroadcast(youtubeURL string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.broadcasts[youtubeURL]; ok {
		return fmt.Errorf("broadcast already exists: %s", youtubeURL)
	}

	port, err := b.getNextPort()
	if err != nil {
		return fmt.Errorf("start broadcast failed to get next port: %w", err)
	}

	frameChan, err := b.StreamService.GetFrameChannel(youtubeURL)
	if err != nil {
		return fmt.Errorf("start broadcast failed to get frame channel: %w", err)
	}

	addr := &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: port,
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("start broadcast failed to listen on port: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	broadcast := &Broadcast{
		YoutubeURL: youtubeURL,
		Port:       port,
		conn:       conn,
		cancel:     cancel,
		ctx:        ctx,
		frameChan:  frameChan,
	}
	b.broadcasts[youtubeURL] = broadcast

	go b.broadcastLoop(broadcast)
	log.Printf("Broadcast started; youtubeURL=%s, port=%d\n", youtubeURL, port)
	return nil
}

func (b *BroadcastService) StopBroadcast(youtubeURL string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	broadcast, ok := b.broadcasts[youtubeURL]
	if !ok {
		return fmt.Errorf("broadcast not found; youtubeURL=%s", youtubeURL)
	}
	broadcast.cancel()
	delete(b.broadcasts, youtubeURL)
	log.Printf("Broadcast stopped; youtubeURL=%s\n", youtubeURL)
	return nil
}

func (b *BroadcastService) GetBroadcast(youtubeURL string) (*Broadcast, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	broadcast, ok := b.broadcasts[youtubeURL]
	if !ok {
		return nil, fmt.Errorf("broadcast not found; youtubeURL=%s", youtubeURL)
	}
	return broadcast, nil
}
func (b *BroadcastService) StopAllActiveBroadcasts() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, broadcast := range b.broadcasts {
		broadcast.cancel()
	}
	b.broadcasts = make(map[string]*Broadcast)
	return nil
}

func (b *BroadcastService) Close() {
	b.StopAllActiveBroadcasts()
	log.Println("Closed broadcast service")
}
