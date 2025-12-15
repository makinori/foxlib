package foxcss

import (
	"context"
	"log/slog"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/cespare/xxhash/v2"
	"github.com/elliotchance/orderedmap/v3"
)

type pageStylesKeyType string
type hashWordsKeyType string

var (
	pageStylesKey pageStylesKeyType = "foxcssPageStyles"
	hashWordsKey  hashWordsKeyType  = "foxcssHashWords"
)

type pageStyles struct {
	classMap *orderedmap.OrderedMap[string, string]
	mutex    sync.Mutex
}

type hashWords struct {
	available []string
	cache     map[string]string
	mutex     sync.Mutex
}

func (hashWords *hashWords) getWord(className string) string {
	hashWords.mutex.Lock()
	defer hashWords.mutex.Unlock()

	word, ok := hashWords.cache[className]
	if ok {
		return word
	}

	seedUint := xxhash.Sum64([]byte(className))
	seedInt := *(*int64)(unsafe.Pointer(&seedUint))
	r := rand.New(rand.NewSource(seedInt))

	i := r.Intn(len(hashWords.available))
	word = hashWords.available[i]

	hashWords.available = slices.Delete(hashWords.available, i, i+1)
	hashWords.cache[className] = word

	return word
}

func InitContext(parent context.Context) context.Context {
	return context.WithValue(
		parent, pageStylesKey,
		&pageStyles{
			classMap: orderedmap.NewOrderedMap[string, string](),
			mutex:    sync.Mutex{},
		},
	)
}

func UseWords(
	parent context.Context, words []string, seed string,
) context.Context {
	hashWords := hashWords{
		cache: map[string]string{},
		mutex: sync.Mutex{},
	}

	for _, word := range words {
		if len(word) == 0 {
			continue
		}

		word = strings.TrimSpace(word)
		word = strings.ToLower(word)
		word = strings.ReplaceAll(word, " ", "-")

		if !slices.Contains(hashWords.available, word) {
			hashWords.available = append(hashWords.available, word)
		}
	}

	return context.WithValue(parent, hashWordsKey, &hashWords)
}

func classNameHash(data []byte) string {
	hash64 := xxhash.Sum64(data)
	hash32 := uint32(hash64>>32) ^ uint32(hash64)
	// c for class i suppose
	return "c" + strconv.FormatUint(uint64(hash32), 36)
}

// returns class name and injects scss into page
func Class(ctx context.Context, scssSnippet string) string {
	if scssSnippet == "" {
		return ""
	}

	pageStyles, ok := ctx.Value(
		pageStylesKey,
	).(*pageStyles)
	if !ok {
		slog.Error("failed to get page scss from context")
		return ""
	}

	// TODO: hash doesnt consider whitespace
	var className = classNameHash([]byte(scssSnippet))

	hashWords, hasHashWords := ctx.Value(hashWordsKey).(*hashWords)
	if hasHashWords {
		className = hashWords.getWord(className)
	}

	pageStyles.mutex.Lock()
	defer pageStyles.mutex.Unlock()

	if pageStyles.classMap.Has(className) {
		return className
	}

	pageStyles.classMap.Set(className, scssSnippet)

	return className
}

func GetPageSCSS(ctx context.Context) string {
	pageStyles, ok := ctx.Value(
		pageStylesKey,
	).(*pageStyles)
	if !ok {
		slog.Error("failed to get page scss from context")
		return ""
	}

	var source string

	classMap := pageStyles.classMap
	for style := classMap.Front(); style != nil; style = style.Next() {
		source += "." + style.Key + "{" + style.Value + "}"
	}

	source = strings.TrimSpace(source)

	return source
}
