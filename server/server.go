package server

import (
	"Twitter/feed"
	"Twitter/queue"
	"encoding/json"
	"log"
	"sync"
)

type Config struct {
	Encoder *json.Encoder // Represents the buffer to encode Responses
	Decoder *json.Decoder // Represents the buffer to decode Requests
	Mode    string        // Represents whether the server should execute
	// sequentially or in parallel
	// If Mode == "s"  then run the sequential version
	// If Mode == "p"  then run the parallel version
	// These are the only values for Version
	ConsumersCount int // Represents the number of consumers to spawn
}

type Response struct {
	Success bool 
	Id int 
}

type FeedResponse struct {
	Id int 
	Feed []feed.SerializablePost
}

//Run starts up the twitter server based on the configuration
//information provided and only returns when the server is fully
// shutdown.
func Run(config Config) {
	feed := feed.NewFeed() 

	if config.Mode == "s"{
		for config.Decoder.More() {
			// Stores the current JSON in the r struct 
			var r queue.Request
			if err := config.Decoder.Decode(&r); err != nil {
				log.Println(err)
				return
			}
			
			if r.Command == "ADD"{
				feed.Add(r.Body, r.Timestamp)
				response := Response{Success: true, 
									  Id: r.ID}
				config.Encoder.Encode(response)
			} else if r.Command == "REMOVE"{
				status := feed.Remove(r.Timestamp)
				response := &Response{Success: status, 
									  Id: r.ID}
				config.Encoder.Encode(response)
			} else if r.Command == "CONTAINS"{
				status := feed.Contains(r.Timestamp)
				response := &Response{Success: status, 
									  Id: r.ID}
				config.Encoder.Encode(response)
			} else if r.Command == "FEED"{
				response := &FeedResponse{Id: r.ID, 
									  Feed: feed.Feed()}
				config.Encoder.Encode(response)
			} else if r.Command == "DONE"{
				break 
			}
		}
	} else {

		var requests* queue.LockFreeQueue  = queue.NewLockFreeQueue();
		
		m := &sync.Mutex{}
		cond := sync.NewCond(m)
		
		var done bool = false 

		var wg sync.WaitGroup

		for i:=0;i<config.ConsumersCount;i++{
			wg.Add(1)
			go consumer(requests, cond, config.Encoder, &done, &wg, feed)
		}

		producer(requests, cond, config.Decoder, &done, &wg, feed) 
	}
}

func producer(requests *queue.LockFreeQueue, cond *sync.Cond, Decoder *json.Decoder, done *bool, wg *sync.WaitGroup, feed feed.Feed){
	
	for (*done == false){
		for Decoder.More() {
			
			var r queue.Request
			if err := Decoder.Decode(&r); err != nil {
				log.Println(err)
				return
			}
			
			if r.Command == "DONE"{
				*done = true 
				requests.Enqueue(&r)
				cond.Broadcast()
				break
			}

			requests.Enqueue(&r)
			cond.Signal()
		}
	}
	wg.Wait() 
}

func consumer(requests *queue.LockFreeQueue, cond *sync.Cond, Encoder *json.Encoder, done *bool, wg *sync.WaitGroup, feed feed.Feed){
	for {
		r := requests.Dequeue();
		if r == nil {
			if *done == true {
				wg.Done()
				return
			}
			cond.L.Lock()
			cond.Wait() 
			cond.L.Unlock()
		} else {

			if r.Command == "ADD"{
				feed.Add(r.Body, r.Timestamp)
				response := Response{Success: true, 
									Id: r.ID}
				Encoder.Encode(response)
			} else if r.Command == "REMOVE"{
				status := feed.Remove(r.Timestamp)
				response := &Response{Success: status, 
									Id: r.ID}
				Encoder.Encode(response)
			} else if r.Command == "CONTAINS"{
				status := feed.Contains(r.Timestamp)
				response := &Response{Success: status, 
									Id: r.ID}
				Encoder.Encode(response)
			} else if r.Command == "FEED"{
				response := &FeedResponse{Id: r.ID, 
									Feed: feed.Feed()}
				Encoder.Encode(response)
			} else if r.Command == "DONE"{
				cond.Broadcast()
			} 
		}
	}
}
