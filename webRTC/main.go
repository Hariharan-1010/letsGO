package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

const (
	webPort = ":8080"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	ws   *websocket.Conn
	peer *webrtc.PeerConnection
}

var clients = make(map[*websocket.Conn]*Client)

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	fmt.Printf("Starting server at http://localhost%s\n", webPort)
	log.Fatal(http.ListenAndServe(webPort, nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	peerConnection, err := createPeerConnection()

	if err != nil {
		log.Println("Failed to create peer connection:", err)
		return
	}

	client := &Client{ws: conn, peer: peerConnection}
	clients[conn] = client

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}

		candidateJSON, _ := json.Marshal(map[string]interface{}{
			"type":      "candidate",
			"candidate": candidate.ToJSON(),
		})
		conn.WriteMessage(websocket.TextMessage, candidateJSON)
	})

	// Handle incoming messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		// Process signalling messages
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println(err)
			continue
		}

		switch msg["type"] {
		case "join":
			log.Printf("%s joined the session", msg["name"])

			offer, err := client.peer.CreateOffer(nil)
			if err != nil {
				log.Println("CreateOffer falied: ", err)
				continue
			}

			offerJSON, _ := json.Marshal(map[string]interface{}{
				"type": "offer",
				"sdp":  offer.SDP,
			})
			conn.WriteMessage(websocket.TextMessage, offerJSON)

		case "offer":
			log.Println("Received offer")
			offer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  msg["sdp"].(string),
			}

			err := client.peer.SetRemoteDescription(offer)
			if err != nil {
				log.Println("SetRemoteDescription Failed: ", err)
				continue
			}

			answer, err := client.peer.CreateAnswer(nil)
			if err != nil {
				log.Println("CreateAnswer falied: ", err)
				continue
			}

			answerJSON, _ := json.Marshal(map[string]interface{}{
				"type": "answer",
				"sdp":  answer.SDP,
			})

			conn.WriteMessage(websocket.TextMessage, answerJSON)

		case "answer":
			log.Println("Received answer")
			answer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  msg["sdp"].(string),
			}
			client.peer.SetRemoteDescription(answer)

		case "candidate":
			log.Println("Received candidate")
			candMap := msg["candidate"].(map[string]interface{})
			candJSON, _ := json.Marshal(candMap)

			var candidate webrtc.ICECandidateInit
			if err := json.Unmarshal(candJSON, &candidate); err != nil {
				log.Println("Failed to unmarshal candidate:", err)
				continue
			}
			client.peer.AddICECandidate(candidate)
		}
	}
}

func createPeerConnection() (*webrtc.PeerConnection, error) {
	// Define ICE servers
	iceServers := []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.1.google.com:19302"},
		},
	}

	// Create a new RTCPeerConnection
	config := webrtc.Configuration{
		ICEServers: iceServers,
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	// Handle ICE connection state changes
	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", state.String())
	})

	return peerConnection, nil
}
