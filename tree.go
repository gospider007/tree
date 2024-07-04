package tree

import (
	"sync"

	"github.com/gospider007/kinds"
)

type Client struct {
	dataStr map[rune]*kinds.Set[string]
	dataLen map[rune]*kinds.Set[int]
	lock    sync.RWMutex
}

func NewClient() *Client {
	return &Client{
		dataStr: map[rune]*kinds.Set[string]{},
		dataLen: map[rune]*kinds.Set[int]{},
	}
}

func (obj *Client) Add(words string) {
	if words == "" {
		return
	}
	wordrunes := []rune(words)
	word_one := wordrunes[0]
	word_str := string(wordrunes[1:])
	word_len := len(wordrunes[1:])
	if !obj.add(word_one, word_str, word_len) {
		obj.lock.Lock()
		obj.dataLen[word_one] = kinds.NewSet(word_len)
		obj.dataStr[word_one] = kinds.NewSet(word_str)
		obj.lock.Unlock()
	}
}
func (obj *Client) add(word_one rune, word_str string, word_len int) bool {
	obj.lock.RLock()
	defer obj.lock.RUnlock()
	lenData, ok := obj.dataLen[word_one]
	if !ok {
		return ok
	}
	lenData.Add(word_len)
	obj.dataStr[word_one].Add(word_str)
	return ok
}
func (obj *Client) Search(wordstr string) map[string]int {
	search_dic := map[string]int{}
	if wordstr == "" {
		return search_dic
	}
	obj.lock.RLock()
	defer obj.lock.RUnlock()
	words := []rune(wordstr)
	words_len := len(words)
	for word_start, word := range words {
		wordLens, ok := obj.dataLen[word]
		if ok {
			last_len := words_len - word_start - 1 //剩余长度=总长度减去-查询过的长度
			for word_len := range wordLens.Map() {
				if word_len > last_len {
					continue
				}
				if qg_str := string(words[word_start+1 : word_start+1+word_len]); obj.dataStr[word].Has(qg_str) {
					searchVal := string(word) + qg_str
					if search_value, search_ok := search_dic[searchVal]; search_ok {
						search_dic[searchVal] = search_value + 1
					} else {
						search_dic[searchVal] = 1
					}
				}
			}
		}
	}
	return search_dic
}
