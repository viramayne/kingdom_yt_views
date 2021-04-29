package main

import "bytes"

func BeautifyNumbers(s string) string {
	var buffer bytes.Buffer
	var n_1 = 2
	var l_1 = len(s) - 1
	str := reverseString(s)

	for i, rune := range str {
		buffer.WriteRune(rune)
		if i%3 == n_1 && i != l_1 {
			buffer.WriteRune(' ')
		}
	}
	return reverseString(buffer.String())
}

func reverseString(s string) string {
	var buffer bytes.Buffer
	var y []byte

	for i := len(s) - 1; i >= 0; i-- {
		y = append(y, byte(s[i]))
	}
	buffer.Write(y)
	str := buffer.String()
	return str
}
