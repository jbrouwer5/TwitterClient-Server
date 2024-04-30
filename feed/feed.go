package feed

import (
	"Twitter/lock"
	"strconv"
)

type SerializablePost struct {
    Body      string  `json:"body"`
    Timestamp float64 `json:"timestamp"`
}

//Feed represents a user's twitter feed
type Feed interface {
	Add(string, float64) 
	Remove(float64) bool
	Contains(float64) bool
	Feed() []SerializablePost
}

//feed is the internal representation of a user's twitter feed (hidden from outside packages)
type feed struct {
	start *post // a pointer to the beginning post
	lock *lock.RWLock // a pointer to a RWLock
}

//post is the internal representation of a post on a user's twitter feed (hidden from outside packages)
type post struct {
	body      string // the text of the post
	timestamp float64  // Unix timestamp of the post
	next      *post  // the next post in the feed
}

//NewPost creates and returns a new post value given its body and timestamp
func newPost(body string, timestamp float64, next *post) *post {
	return &post{body, timestamp, next}
}

//NewPost creates and returns a new post value given its body and timestamp
func newLock() *lock.RWLock {
	return lock.NewRWLock()
}

//NewFeed creates a empy user feed
func NewFeed() Feed {
	return &feed{start: nil, lock: newLock()}
}

func (f *feed) PrintFeed() {
	pointer := f.start
	for pointer != nil && pointer.next != nil{
		print(strconv.FormatFloat(pointer.timestamp, 'f', -1, 64) + ", ")
		pointer = pointer.next
	}
}

// Add inserts a new post to the feed. The feed is always ordered by the timestamp where
// the most recent timestamp is at the beginning of the feed followed by the second most
// recent timestamp, etc. You may need to insert a new post somewhere in the feed because
// the given timestamp may not be the most recent.
func (f *feed) Add(body string, timestamp float64) {
	f.lock.Lock() 

	if f.start == nil{
		f.start = newPost(body, timestamp, nil)
		f.lock.Unlock() 
		return 
	} else if f.start.timestamp < timestamp {
		f.start = newPost(body, timestamp, f.start)
		f.lock.Unlock() 
		return 
	}

	pointer := f.start
	for pointer != nil && pointer.next != nil && pointer.next.timestamp >= timestamp{
		pointer = pointer.next
	}

	if pointer.next == nil {
		pointer.next = newPost(body, timestamp, nil)
	} else {
		tmp := pointer.next 
		pointer.next = newPost(body, timestamp, tmp)
	}
	f.lock.Unlock() 
}

// Remove deletes the post with the given timestamp. If the timestamp
// is not included in a post of the feed then the feed remains
// unchanged. Return true if the deletion was a success, otherwise return false
func (f *feed) Remove(timestamp float64) bool {
	f.lock.Lock() 

	if f.start == nil{
		f.lock.Unlock() 
		return false
	} else if f.start.timestamp == timestamp {
		f.start = f.start.next
		f.lock.Unlock() 
		return true 
	}

	pointer := f.start
	for pointer.next != nil && pointer.next.timestamp > timestamp{
		pointer = pointer.next
	}

	if pointer.next == nil{
		f.lock.Unlock() 
		return false 
	}

	if pointer.next.timestamp == timestamp{
		tmp := pointer.next 
		pointer.next = tmp.next
		f.lock.Unlock() 
		return true 
	}

	f.lock.Unlock() 
	return false 
}

// Contains determines whether a post with the given timestamp is
// inside a feed. The function returns true if there is a post
// with the timestamp, otherwise, false.
func (f *feed) Contains(timestamp float64) bool {
	f.lock.RLock() 
	if f.start == nil{
		f.lock.RUnlock() 
		return false
	} else if f.start.timestamp == timestamp {
		f.lock.RUnlock() 
		return true 
	}

	pointer := f.start
	for pointer.next != nil && pointer.next.timestamp > timestamp{
		pointer = pointer.next
	}

	if pointer.next == nil {
		f.lock.RUnlock() 
		return false 
	}
	
	if pointer.next.timestamp == timestamp{
		f.lock.RUnlock() 
		return true 
	}

	f.lock.RUnlock() 
	return false 
}

// Contains determines whether a post with the given timestamp is
// inside a feed. The function returns true if there is a post
// with the timestamp, otherwise, false.
func (f *feed) Feed() []SerializablePost {
	var posts []SerializablePost

	f.lock.RLock() 
	for p := f.start; p != nil; p = p.next {
		post := SerializablePost{
			Body:      p.body,
			Timestamp: p.timestamp,
		}
		posts = append(posts, post)
	}
	f.lock.RUnlock() 
	return posts
}


