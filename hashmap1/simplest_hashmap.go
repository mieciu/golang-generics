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
}

// This is the simplest hashmap implementation
// in case of hash collision it just doubles its size
// until keyspace won't end up in a collision

type HashMap[K comparable, V any] struct {
	capacity int64
	entries  []*KVPair[K, V]
}

func (m *HashMap[K, V]) get(key K) *V {
	hashedKey := m.hash(key)
	if m.entries[hashedKey] != nil {
		return &m.entries[hashedKey].Value
	} else {
		return nil
	}
}

func (m *HashMap[K, V]) set(key K, value V) {
	hashedKey := m.hash(key)
	if m.entries[hashedKey] == nil {
		kvPairToInsert := KVPair[K, V]{Key: key, Value: value}
		m.entries[hashedKey] = &kvPairToInsert
	} else {
		if m.entries[hashedKey].Key == key {
			m.entries[hashedKey].Value = value
		} else {
			m.rehash(key)
			m.set(key, value)
		}
	}
}

// Rehash map so that newKey won't cause a collision
func (m *HashMap[K, V]) rehash(newKey K) {
	oldKeyspace := make([]K, len(m.entries))
	newKeyspace := append(oldKeyspace, newKey)
	for ok := true; ok; ok = m.noCollidingHashes(newKeyspace) {
		m.capacity = m.capacity * 2
	}
	oldEntries := m.entries

	m.entries = make([]*KVPair[K, V], m.capacity)
	for _, oldEntry := range oldEntries {
		if oldEntry != nil {
			m.set(oldEntry.Key, oldEntry.Value)
		}
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
	m.entries[hashedKey] = nil
}

func MakeHashMap[K comparable, V any]() *HashMap[K, V] {
	defaultCapacity := 4
	return &HashMap[K, V]{
		capacity: int64(defaultCapacity),
		entries:  make([]*KVPair[K, V], defaultCapacity),
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
	println(myHashmap.get("sdf"))
	println(myHashmap.get("asdf"))
	println(myHashmap.get("asdfs"))
	println(myHashmap.get("asd2342342f"))
	println(myHashmap.get("non-existent"))

	println(myHashmap)
}
