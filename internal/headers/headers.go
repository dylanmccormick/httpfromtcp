package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	H := make(Headers)
	return H
}

func (h Headers) Get(key string) (value string, err error) {
	if _, ok := h[strings.ToLower(key)]; ok {
		return h[strings.ToLower(key)], nil
	}
	return "", fmt.Errorf("Key not found... nerd")
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	data_string := string(data)

	if strings.HasPrefix(data_string, "\r\n") {
		return 2, true, nil
	}

	endIndex := FindCRLF(data_string)
	if endIndex == -1 {
		return 0, false, nil
	}
	parsedLength := endIndex + 2

	line := data_string[0:endIndex]
	parts := strings.SplitN(line, ":", 2)

	if len(parts) < 2 {
		return 0, false, fmt.Errorf("Invalid header provided")
	}

	left := parts[0]
	right := parts[1]
	rightTrim := strings.Trim(right, " \t")

	leftTrim, err := checkFieldName(left)
	if err != nil {
		return 0, false, err
	}

	input := standardizeFieldName(leftTrim)

	if _, ok := h[input]; !ok {
		h[input] = rightTrim
	} else {
		h[input] = h[input] + ", " + rightTrim
	}

	return parsedLength, false, nil
}

func standardizeFieldName(fieldName string) string {
	return strings.ToLower(fieldName)
}

func checkFieldName(fieldName string) (string, error) {
	if len(fieldName) < 1 {
		return "", fmt.Errorf("Invalid Header: Field name is too short. Must be at least one character")
	}
	lastChar := fieldName[len(fieldName)-1]
	if lastChar == ' ' || lastChar == '\t' {
		return "", fmt.Errorf("Whitespace found after field name and before ':'. Invalid header")
	}

	newString := strings.Trim(fieldName, " \t")
	approvedMap := map[rune]struct{}{
		'!':  {},
		'#':  {},
		'$':  {},
		'%':  {},
		'&':  {},
		'\'': {},
		'*':  {},
		'+':  {},
		'-':  {},
		'.':  {},
		'^':  {},
		'_':  {},
		'`':  {},
		'|':  {},
		'~':  {},
	}

	for i := 48; i <= 57; i++ {
		approvedMap[rune(i)] = struct{}{}
	}
	for i := 65; i <= 90; i++ {
		approvedMap[rune(i)] = struct{}{}
	}
	for i := 97; i <= 122; i++ {
		approvedMap[rune(i)] = struct{}{}
	}

	for _, c := range newString {
		if _, ok := approvedMap[c]; !ok {
			return "", fmt.Errorf("Invalid Character found: %c", c)
		}
	}

	return newString, nil

}

func FindCRLF(msg string) int {
	for i := 0; i < len(msg)-1; i++ {
		if msg[i] == '\r' && msg[i+1] == '\n' {
			return i
		}
	}
	return -1
}
