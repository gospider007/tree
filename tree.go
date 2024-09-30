package tree

import (
	"encoding/json"
	"os"
	"slices"
	"sync"

	"github.com/gospider007/kinds"
	"github.com/gospider007/tools"
	"github.com/klauspost/compress/zstd"
)

type dataLenValue struct {
	maxWord    bool
	value      *kinds.Set[int]
	isSort     bool
	orderValue []int
}

func (obj *dataLenValue) Add(value int) {
	obj.isSort = false
	obj.value.Add(value)
}
func (obj *dataLenValue) Array() []int {
	if obj.isSort {
		return obj.orderValue
	}
	values := obj.value.Array()
	slices.Sort(values)
	if !obj.maxWord {
		slices.Reverse(values)
	}
	obj.orderValue = values
	obj.isSort = true
	return values
}

type Client struct {
	maxWord bool
	minLen  int
	dataStr map[rune]*kinds.Set[string]
	dataLen map[rune]*dataLenValue
	lock    sync.RWMutex
}
type ClientOption struct {
	MaxWord bool
	MinLen  int
}

func NewClient(option ...ClientOption) *Client {
	var opt ClientOption
	if len(option) > 0 {
		opt = option[0]
	}
	return &Client{
		maxWord: opt.MaxWord,
		minLen:  opt.MinLen,
		dataStr: make(map[rune]*kinds.Set[string]),
		dataLen: make(map[rune]*dataLenValue),
	}
}

func (obj *Client) Add(words string) {
	if words == "" {
		return
	}
	wordrunes := []rune(words)
	word_len := len(wordrunes[1:])
	if word_len < obj.minLen {
		return
	}
	word_one := wordrunes[0]
	word_str := string(wordrunes[1:])

	if !obj.add(word_one, word_str, word_len) {
		obj.lock.Lock()
		obj.dataLen[word_one] = &dataLenValue{value: kinds.NewSet(word_len), maxWord: obj.maxWord}
		obj.dataStr[word_one] = kinds.NewSet(word_str)
		obj.lock.Unlock()
	}
}
func (obj *Client) get(word_one rune) (*dataLenValue, bool) {
	obj.lock.RLock()
	lenData, ok := obj.dataLen[word_one]
	obj.lock.RUnlock()
	return lenData, ok
}
func (obj *Client) add(word_one rune, word_str string, word_len int) bool {
	lenData, ok := obj.get(word_one)
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
	for word_start := 0; word_start < words_len; word_start++ {
		word := words[word_start]
		wordLens, ok := obj.dataLen[word]
		if ok {
			last_len := words_len - word_start - 1 //剩余长度=总长度减去-查询过的长度
			for _, word_len := range wordLens.Array() {
				if word_len > last_len {
					if obj.maxWord {
						break
					} else {
						continue
					}
				}
				if qg_str := string(words[word_start+1 : word_start+1+word_len]); obj.dataStr[word].Has(qg_str) {
					searchVal := string(word) + qg_str
					if search_value, search_ok := search_dic[searchVal]; search_ok {
						search_dic[searchVal] = search_value + 1
					} else {
						search_dic[searchVal] = 1
					}
					word_start += word_len
					break
				}
			}
		}
	}
	return search_dic
}

type DataLenValueClone struct {
	MaxWord    bool             `json:"MaxWord"`
	Value      map[int]struct{} `json:"Value"`
	IsSort     bool             `json:"IsSort"`
	OrderValue []int            `json:"OrderValue"`
}
type ClientClone struct {
	MaxWord bool                         `json:"MaxWord"`
	MinLen  int                          `json:"MinLen"`
	DataStr map[rune]map[string]struct{} `json:"DataStr"`
	DataLen map[rune]DataLenValueClone   `json:"DataLen"`
}

func (obj *Client) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zs, err := zstd.NewWriter(f)
	if err != nil {
		return err
	}
	defer zs.Close()
	// encoder := gob.NewEncoder(zs)
	encoder := json.NewEncoder(zs)
	dataStr := make(map[rune]map[string]struct{})
	for k, v := range obj.dataStr {
		dataStr[k] = v.Map()
	}
	dataLen := make(map[rune]DataLenValueClone)
	for k, v := range obj.dataLen {
		dataLen[k] = DataLenValueClone{
			MaxWord:    v.maxWord,
			Value:      v.value.Map(),
			IsSort:     v.isSort,
			OrderValue: v.orderValue,
		}
	}
	return encoder.Encode(ClientClone{
		MaxWord: obj.maxWord,
		MinLen:  obj.minLen,
		DataStr: dataStr,
		DataLen: dataLen,
	})
}
func Load(path string) (*Client, error) {
	f2, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f2.Close()
	f, err := tools.CompressionDecode(f2, "zstd")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var clientClone ClientClone
	decoder := json.NewDecoder(f)
	// err = gob.NewDecoder(f).Decode(&clientClone)
	err = decoder.Decode(&clientClone)
	if err != nil {
		return nil, err
	}
	var client Client
	client.maxWord = clientClone.MaxWord
	client.minLen = clientClone.MinLen
	client.dataStr = make(map[rune]*kinds.Set[string])
	client.dataLen = make(map[rune]*dataLenValue)
	for k, v := range clientClone.DataStr {
		client.dataStr[k] = kinds.NewSetWithMap(v)
	}
	for k, v := range clientClone.DataLen {
		client.dataLen[k] = &dataLenValue{
			maxWord:    v.MaxWord,
			value:      kinds.NewSetWithMap(v.Value),
			isSort:     v.IsSort,
			orderValue: v.OrderValue,
		}
	}
	return &client, nil
}
