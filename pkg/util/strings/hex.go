package strings

import (
	"fmt"
	"strings"
)

// HexFormat format buffer like hex editor
func HexFormat(b []byte) string {
	n := len(b)
	rowCount := n / 16
	sb := strings.Builder{}
	sbASCII := make([]byte, 16)
	for row := 0; row < rowCount; row++ {
		for col := 0; col < 16; col++ {
			c := b[row*16+col]
			sb.WriteString(fmt.Sprintf("%02x ", c))
			if c >= 0x20 && c <= 0x7E {
				sbASCII[col] = c
			} else {
				sbASCII[col] = 0x20
			}
		}
		sb.WriteString(" | ")
		sb.WriteString(string(sbASCII))
		sb.WriteString("\n")
	}
	if n%16 != 0 {
		// the last line
		cnt := n % 16
		for col := 0; col < cnt; col++ {
			c := b[rowCount*16+col]
			sb.WriteString(fmt.Sprintf("%02x ", c))
			if c >= 0x20 && c <= 0x7E {
				sbASCII[col] = c
			} else {
				sbASCII[col] = 0x20
			}
		}
		for col := cnt; col < 16; col++ {
			sb.WriteString("   ")
			sbASCII[col] = 0x20
		}
		sb.WriteString(" | ")
		sb.WriteString(string(sbASCII))
		sb.WriteString("\n")
	}
	return sb.String()
}

// ToHexString convert to golang string, not ascii char convert to \xXX
func ToHexString(buf []byte) string {
	sb := strings.Builder{}
	sb.WriteString("s:=\"")
	for _, c := range buf {
		if c == 0x22 || c == 0x5c || c == 60 {
			sb.WriteString(fmt.Sprintf("\\x%02x", c))
		} else if c >= 0x20 && c <= 0x7E {
			sb.WriteByte(c)
		} else {
			sb.WriteString(fmt.Sprintf("\\x%02x", c))
		}
	}
	sb.WriteString("\"")
	return sb.String()
}
