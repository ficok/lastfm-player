package main

import (
	"errors"
	"fmt"
)

type DoubleList struct {
	first *Node
	last  *Node
	empty bool
	size  int
}

type Node struct {
	prev     *Node
	idx      int
	priority bool
	next     *Node
}

func pushFront(pair Pair) {
	// create a new node
	var oldFirst *Node
	// all this to not dereference nil
	if downloadQueue.first == nil {
		oldFirst = nil
	} else {
		oldFirst = downloadQueue.first
	}
	// setting first as the new node
	downloadQueue.first = &Node{
		// the first node has no prev
		prev:     nil,
		idx:      pair.idx,
		priority: pair.priority,
		next:     oldFirst,
		/*
			when pushing to the front, the old first node
			becomes the second node
			if the queue was empty before this, the oldFirst is nil
			and the first node will have nil for both prev and
			next
		*/
	}
	// if this is the first node, change the value of empty and
	// set the last node to point to the first node
	if downloadQueue.empty {
		// in case something went wrong
		if downloadQueue.first != nil {
			downloadQueue.last = downloadQueue.first
			downloadQueue.empty = false
		} else {
			fmt.Println("ERROR[pushFront]: did not create a new node")
		}
	}
	// increment size
	downloadQueue.size++
}

func pushBack(pair Pair) {
	// create a new node
	var oldLast *Node
	// all this to not dereference nil
	if downloadQueue.last == nil {
		oldLast = nil
	} else {
		oldLast = downloadQueue.last
	}
	// setting last as the new node
	downloadQueue.last = &Node{
		// the last node has no next
		prev:     oldLast,
		idx:      pair.idx,
		priority: pair.priority,
		next:     nil,
		/*
			when pushing to the back, the old last node
			becomes the second last node
			if the queue was empty before this, the oldLast is nil
			and the first node will have nil for both prev and
			next
		*/
	}

	// if this is the first node, change the value of empty and
	// set the first node to point to the last node
	if downloadQueue.empty {
		// in case something went wrong
		if downloadQueue.last != nil {
			downloadQueue.first = downloadQueue.last
			downloadQueue.empty = false
		} else {
			fmt.Println("ERROR[pushBack]: did not create a new node")
		}
	}
	// increment size
	downloadQueue.size++
}

func readFront() (Pair, error) {
	pair := Pair{idx: -1, priority: false}
	if downloadQueue.first == nil {
		return pair, errors.New("ERROR[readFront]: cannot read empty queue")
	}

	pair.idx = downloadQueue.first.idx
	pair.priority = downloadQueue.first.priority

	// fmt.Println("INFO[readFront]: Pair from readFront:", pair)

	return pair, nil
}

func readBack() (Pair, error) {
	pair := Pair{idx: -1, priority: false}
	if downloadQueue.empty {
		return pair, errors.New("ERROR[readBack]: cannot read from an empty queue")
	}

	pair.idx = downloadQueue.last.idx
	pair.priority = downloadQueue.last.priority

	return pair, nil
}

// don't forget to set empty to true if it's emptied this way
func popFront() error {
	fmt.Println("#### DEBUG[popFront]: print downloadQueue", downloadQueue)
	if downloadQueue.empty {
		return errors.New("ERROR[popFront]: cannot pop from an empty queue")
	}

	// set the first to first.next
	if downloadQueue.first.next == nil {
		// if nil, this was the last node and the queue is now empty
		downloadQueue.first = nil
		downloadQueue.last = nil
		downloadQueue.empty = true
		downloadQueue.size = 0

		return nil
	}

	downloadQueue.first = downloadQueue.first.next

	fmt.Println("#### DEBUG[popFront]: print downloadQueue", downloadQueue)
	fmt.Println("#### DEBUG[popFront]: print downloadQueue.first", downloadQueue.first)
	fmt.Println("#### DEBUG[popFront]: print downloadQueue.first.prev", downloadQueue.first.prev)

	// set the first.prev to nil
	if downloadQueue.first.prev != nil {
		downloadQueue.first.prev = nil
	}

	// decrement size
	downloadQueue.size--

	return nil
}
func popBack() error {
	if downloadQueue.empty {
		return errors.New("ERROR[popBack]: cannot pop from an empty queue")
	}

	// set the last to last.prev
	if downloadQueue.last.prev == nil {
		// if nil, this was the last node and the queue is now empty
		downloadQueue.first = nil
		downloadQueue.last = nil
		downloadQueue.empty = true
		downloadQueue.size = 0

		return nil
	}

	downloadQueue.last = downloadQueue.last.prev

	// set the last.next to nil
	if downloadQueue.last.next != nil {
		downloadQueue.last.next = nil
	}

	// decrement size
	downloadQueue.size--

	return nil
}

func isEmpty() bool {
	return downloadQueue.empty
}
