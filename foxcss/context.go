package foxcss

import (
	"context"
	"errors"
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

var (
	pageStylesKey pageStylesKeyType = "foxcssPageStyles"
)

type hashWords struct {
	available []string
	cache     map[string]string
	mutex     sync.Mutex
}

type pageStyles struct {
	classMap    *orderedmap.OrderedMap[string, string]
	mutex       sync.Mutex
	hashWords   *hashWords
	classPrefix string
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

func InitContext(ctx context.Context, classPrefix string) context.Context {
	return context.WithValue(
		ctx, pageStylesKey,
		&pageStyles{
			classMap:    orderedmap.NewOrderedMap[string, string](),
			mutex:       sync.Mutex{},
			classPrefix: classPrefix,
		},
	)
}

func UseWords(
	ctx context.Context, words []string, seed string,
) error {
	pageStyles, ok := ctx.Value(
		pageStylesKey,
	).(*pageStyles)
	if !ok {
		return errors.New("page styles not found in context")
	}

	pageStyles.hashWords = &hashWords{
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

		if !slices.Contains(pageStyles.hashWords.available, word) {
			pageStyles.hashWords.available = append(
				pageStyles.hashWords.available, word,
			)
		}
	}

	return nil
}

// need to still prefix with a-z
func classNameHash(data []byte) string {
	hash64 := xxhash.Sum64(data)
	hash32 := uint32(hash64>>32) ^ uint32(hash64)
	return strconv.FormatUint(uint64(hash32), 36)
}

// returns class name and injects snipper into page styles
func Class(ctx context.Context, snippet string) string {
	if snippet == "" {
		return ""
	}

	pageStyles, ok := ctx.Value(
		pageStylesKey,
	).(*pageStyles)
	if !ok {
		slog.Error("failed to get page styles from context")
		return ""
	}

	css := preprocess(snippet)
	var className = classNameHash([]byte(css))

	if pageStyles.hashWords != nil {
		className = pageStyles.hashWords.getWord(className)
	}

	if pageStyles.classPrefix == "" {
		if pageStyles.hashWords == nil {
			// c for class
			className = "c" + className
		}
	} else {
		className = pageStyles.classPrefix + className
	}

	pageStyles.mutex.Lock()
	defer pageStyles.mutex.Unlock()

	if pageStyles.classMap.Has(className) {
		return className
	}

	pageStyles.classMap.Set(className, css)

	return className
}

func GetPageCSS(ctx context.Context) string {
	pageStyles, ok := ctx.Value(
		pageStylesKey,
	).(*pageStyles)
	if !ok {
		slog.Error("failed to get page css from context")
		return ""
	}

	var css string

	classMap := pageStyles.classMap
	for style := classMap.Front(); style != nil; style = style.Next() {
		css += strings.ReplaceAll(style.Value, "&", "."+style.Key)
	}

	return css
}
