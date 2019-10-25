package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pion/webrtc"
)

func main() {
	var (
		api *webrtc.API
		timeout = 5 * time.Second
	)

	// Create a new API with Trickle ICE enabled.
	// This SettingEngine allows non-standard WebRTC behavior.
	s := webrtc.SettingEngine{}
	s.SetTrickle(true)
	api = webrtc.NewAPI(webrtc.WithSettingEngine(s))

	var wg sync.WaitGroup

	ICEServers := []webrtc.ICEServer{
		webrtc.ICEServer{
			URLs: []string{os.Args[1]},
		},
	}

	g, err := api.NewICEGatherer(webrtc.ICEGatherOptions{
		ICEServers:      ICEServers,
		ICEGatherPolicy: webrtc.ICETransportPolicyAll,
	})
	if err != nil {
		fmt.Println("ERROR api.NewICEGatherer: ", err)
		return
	}
	defer g.Close()

	var (
		startTime time.Time
		complete = make(chan struct{})
	)

	g.OnStateChange(func(s webrtc.ICEGathererState) {
		if s == webrtc.ICEGathererStateComplete {
			complete <- struct{}{}
		}
	})

	g.OnLocalCandidate(func(c *webrtc.ICECandidate) {
		wg.Add(1)
		defer wg.Done()

		if c == nil {
			return
		}

		candidatesJSON, err := json.Marshal(c.ToJSON())
		if err != nil {
			fmt.Println("ERROR json.Marshal: ", err)
		}
		fmt.Printf("time: %.6fs\n%s\n", time.Since(startTime).Seconds(), string(candidatesJSON))
	})
	
	fmt.Println("gathering...")
	startTime = time.Now()
	g.Gather()

	select {
	case <- time.After(timeout):
		fmt.Println("timeout!")

	case <- complete:
		time.Sleep(100*time.Millisecond) // wait some time for entering OnLocalCandidate callback
		wg.Wait()
		fmt.Println("gathering complete!")
	}
}