package articles

import (
	"bytes"
	"regexp"

	"github.com/kljensen/snowball"
	"golang.org/x/text/unicode/norm"
)

var (
	// words that are single digit or that dont contain included characters, except "-" in middle ex. mother-in-law
	words_re            = regexp.MustCompile(`([^\s.,_<>[\]!%+^?#|;'$@&\-=*)(/"\\=` + "`" + `:][^\s.,_<>[\]!%?+^#|;'$@&=*)(/"\\=` + "`" + `:]{0,}[^\s.,\-_<>[\]!%?+^#|;'$@&=*)(/"\\=` + "`" + `:])|(\d)`)
	words_no_numbers_re = regexp.MustCompile(`([^\s.,_<>[\]!%+^?#|;\d'$@&\-=*)(/"\\=` + "`" + `:][^\s.,_<>[\]!%?+^#|;\d'$@&=*)(/"\\=` + "`" + `:]{0,}[^\s.,\-_<>[\]!%?+^#|;\d'$@&=*)(/"\\=` + "`" + `:])`)
	stopwords_re        = regexp.MustCompile(`\b(i|me|my|myself|we|our|ours|ourselves|you|your|yours|yourself|yourselves|he|him|his|himself|she|her|hers|herself|it|its|itself|they|them|their|theirs|themselves|what|which|who|whom|this|that|these|those|am|is|are|was|were|be|been|being|have|has|had|having|do|does|did|doing|a|an|the|and|but|if|or|because|as|until|while|of|at|by|for|with|about|against|between|into|through|during|before|after|above|below|to|from|up|down|in|out|on|off|over|under|again|further|then|once|here|there|when|where|why|how|all|any|both|each|few|kmore|most|other|some|such|no|nor|not|only|own|same|so|than|too|very|s|t|can|will|just|don|should|now)\b`)
)

func stemAndRemoveStopWordsString(str string) (string, error) {
	var result []byte
	content := []byte(str)
	content = norm.NFC.Bytes(content)
	content = bytes.ToLower(content)
	words := words_no_numbers_re.FindAll(content, -1)
	for _, word := range words {
		if !stopwords_re.Match(word) {
			word, err := snowball.Stem(string(word), "english", false)
			if err != nil {
				return "", err
			}
			result = append(result, word...)
			result = append(result, ' ')
		}
	}
	return string(result), nil
}
