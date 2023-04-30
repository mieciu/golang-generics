package main

import (
	bytes2 "bytes"
	"crypto/sha256"
	"encoding/gob"
	"math/big"
)

type KVPair[K comparable, V any] struct {
	Key   K
	Value V
	Next  *KVPair[K, V]
}

// This is the simple, but more sophisticated hashmap implementation
// In case of hash collision buckets form a linked-list.

type HashMap[K comparable, V any] struct {
	capacity int64
	buckets  []*KVPair[K, V]

	listLen         int // tracking length of linked list when running set() operation
	rehashThreshold int // when bucket contains this amount of KVPairs, whole Hashmap is going to be rehashed
}

func (m *HashMap[K, V]) get(key K) *V {
	hashedKey := m.hash(key)
	for pointer := m.buckets[hashedKey]; pointer != nil; pointer = pointer.Next {
		if pointer.Key == key {
			return &pointer.Value
		}
	}
	return nil
}

func (m *HashMap[K, V]) resetListLen() {
	m.listLen = 0
}

func (m *HashMap[K, V]) set(key K, value V) {
	defer m.resetListLen()
	hashedKey := m.hash(key)
	kvPairToInsert := KVPair[K, V]{Key: key, Value: value, Next: nil}
	if m.buckets[hashedKey] == nil {
		m.buckets[hashedKey] = &kvPairToInsert
	} else {
		for pointer := m.buckets[hashedKey]; pointer != nil; pointer = pointer.Next {
			m.listLen++
			if pointer.Key == key { // in place update of value
				pointer.Value = value
				break
			}
			if pointer.Next == nil {
				pointer.Next = &kvPairToInsert
			}
			if m.listLen >= m.rehashThreshold {
				m.rehash()
			}
		}
	}
}

// not efficient at all but ..
func (m *HashMap[K, V]) rehash() {
	var allElements []KVPair[K, V]
	for _, bucket := range m.buckets {
		node := bucket
		for node != nil {
			allElements = append(allElements, *node)
			node = node.Next
		}
	}
	keyspace := make([]K, len(allElements))
	for _, entry := range allElements {
		keyspace = append(keyspace, entry.Key)
	}

	for ok := true; ok; ok = m.noCollidingHashes(keyspace) {
		m.capacity = m.capacity * 2
		println("need to grow cap to ", m.capacity)
	}
	m.buckets = make([]*KVPair[K, V], m.capacity)

	for _, entry := range allElements {
		m.set(entry.Key, entry.Value)
	}
}

func (m *HashMap[K, V]) noCollidingHashes(keyspace []K) bool {
	allHashes := make([]int, len(keyspace))
	for i, key := range keyspace {
		allHashes[i] = m.hash(key)
	}
	return len(allHashes) < len(keyspace)
}

func (m *HashMap[K, V]) remove(key K) {
	hashedKey := m.hash(key)
	if m.buckets[hashedKey] == nil {
		return
	}
	if m.buckets[hashedKey].Key == key { // key is in HEAD
		m.buckets[hashedKey] = m.buckets[hashedKey].Next
		return
	}
	prev := m.buckets[hashedKey]
	curr := m.buckets[hashedKey].Next
	for curr != nil {
		if curr.Key == key {
			prev.Next = curr.Next
			return
		}
		prev = prev.Next
		curr = curr.Next
	}
}

func MakeHashMap[K comparable, V any]() *HashMap[K, V] {
	defaultCapacity := 4
	defaultRehashThreshold := 2
	return &HashMap[K, V]{
		capacity:        int64(defaultCapacity),
		buckets:         make([]*KVPair[K, V], defaultCapacity),
		rehashThreshold: defaultRehashThreshold,
	}
}

func (m *HashMap[K, V]) hash(key K) int {
	var buffer bytes2.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(key); err != nil {
		panic(err)
	}
	hashedKeyBytes := sha256.Sum256(buffer.Bytes())
	var bigInt big.Int
	bigInt.SetBytes(hashedKeyBytes[:])
	hashAsInteger := bigInt.Int64()
	hashAfterModulo := int(hashAsInteger % m.capacity)
	if hashAfterModulo < 0 {
		return -hashAfterModulo
	}
	return hashAfterModulo
}

func main() {
	myHashmap := MakeHashMap[string, int]()
	println(myHashmap.hash("sdf"))
	println(myHashmap.hash("asdf"))
	println(myHashmap.hash("asdfs"))
	println(myHashmap.hash("asd2342342f"))
	println("-----------------------------")
	myHashmap.set("sdf", 1)
	myHashmap.set("asdf", 2)
	myHashmap.set("asdfs", 3)
	myHashmap.set("asd2342342f", 4)

	myHashmap.set("sdf2222222", 10)
	myHashmap.set("asdf2222222", 20)
	myHashmap.set("asdfs2222222", 30)
	myHashmap.set("asd2342342f2222222", 40)
	println("-----------------------------")
	println(*myHashmap.get("sdf"))
	println(*myHashmap.get("asdf"))
	println(*myHashmap.get("asdf2222222"))
	println(*myHashmap.get("asd2342342f"))

	println(myHashmap)
}
