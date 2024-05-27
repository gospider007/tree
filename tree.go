package tree

import (
	"sort"
	"sync"

	"github.com/gospider007/kinds"
)

type Client struct {
	dataStr      map[rune]*kinds.Set[string]
	dataLen      map[rune]*kinds.Set[int]
	dataOrdLen   map[rune][]int
	dataSortKeys *kinds.Set[rune]
	lock         sync.Mutex
}

func NewClient() *Client {
	return &Client{
		dataStr:      map[rune]*kinds.Set[string]{},
		dataLen:      map[rune]*kinds.Set[int]{},
		dataOrdLen:   map[rune][]int{},
		dataSortKeys: kinds.NewSet[rune](),
	}
}

func (obj *Client) Add(words string) {
	if words == "" {
		return
	}
	obj.lock.Lock()
	defer obj.lock.Unlock()
	wordrunes := []rune(words)
	word_one := wordrunes[0]
	wordrune_str := wordrunes[1:]
	lenData, ok := obj.dataLen[word_one]
	if ok {
		lenData.Add(len(wordrune_str))
		obj.dataStr[word_one].Add(string(wordrune_str))
	} else {
		obj.dataLen[word_one] = kinds.NewSet(len(wordrune_str))
		obj.dataStr[word_one] = kinds.NewSet(string(wordrune_str))
	}
	obj.dataSortKeys.Add(word_one)
}
func (obj *Client) sort() {
	if obj.dataSortKeys.Len() == 0 {
		return
	}
	for k := range obj.dataSortKeys.Map() {
		vvs := obj.dataLen[k].Array()
		sort.Slice(vvs, func(i, j int) bool {
			return vvs[i] > vvs[j]
		})
		obj.dataOrdLen[k] = vvs
	}
	obj.dataSortKeys.ReSet()
}
func (obj *Client) Search(wordstr string) map[string]int {
	search_dic := map[string]int{}
	if wordstr == "" {
		return search_dic
	}
	obj.lock.Lock()
	defer obj.lock.Unlock()
	obj.sort()
	words := []rune(wordstr)
	words_len := len(words)
	for word_start, word := range words {
		wordLens, ok := obj.dataOrdLen[word]
		if ok {
			last_len := words_len - word_start - 1 //剩余长度=总长度减去-查询过的长度
			for _, word_len := range wordLens {
				if word_len > last_len {
					continue
				}
				qg_str := string(words[word_start+1 : word_start+1+word_len])
				if obj.dataStr[word].Has(qg_str) {
					searchVal := string(word) + qg_str
					search_value, search_ok := search_dic[searchVal]
					if search_ok {
						search_dic[searchVal] = search_value + 1
					} else {
						search_dic[searchVal] = 1
					}
					break
				}
			}
		}
	}
	return search_dic
}
