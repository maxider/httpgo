package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

const SEPERATOR = "\r\n"

func (h Headers) Get(key string) (value string) {
	return h[strings.ToLower(key)]
}

func (h Headers) GetContentLength() (int, error) {
	value := h["content-length"]
	if value == "" {
		return -1, nil
	}

	cl, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return cl, nil
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	countRead := 0

	for {
		i := bytes.Index(data, []byte(SEPERATOR))
		if i == -1 {
			break
		}

		countRead += i + len(SEPERATOR)
		if i == 0 {
			return countRead, true, nil
		}

		fieldLine := data[:i]

		name, value, err := parseFieldLine(fieldLine)
		if err != nil {
			return 0, false, err
		}
		name = strings.ToLower(name)

		isKeyValid, errRune := validateKey(name)
		if !isKeyValid {
			return 0, false, fmt.Errorf("invalid char in header name: %c", *errRune)
		}

		if _, hasKey := h[name]; hasKey {
			h[name] = h[name] + ", " + value
		} else {
			h[name] = value
		}
		data = data[i+len(SEPERATOR):]
	}

	return countRead, false, nil
}

func validateKey(name string) (bool, *rune) {
	validSpecialChars := []rune{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

	for _, r := range name {
		if !unicode.IsLetter(r) && !slices.Contains(validSpecialChars, r) && !unicode.IsNumber(r) {
			return false, &r
		}
	}

	return true, nil
}

func parseFieldLine(data []byte) (string, string, error) {
	parts := bytes.SplitN(data, []byte{':'}, 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed fieldLine")
	}

	key := bytes.TrimPrefix(parts[0], []byte(" "))
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(key, []byte(" ")) {
		return "", "", fmt.Errorf("maleformed key")
	}

	return string(key), string(value), nil
}
