package directEmail

import (
	"time"
	"fmt"
	"bytes"
	"encoding/base64"
	"math/rand"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
	"path/filepath"
	"reflect"
	"unsafe"
)

type Data struct {
	ExtraHeader         []string
	From string
	To     string
	Subject             string
	Html                string //ToDo add Text
	Attachments         []string
	raw                 []byte
}

func (self *Data) get() []byte {
	return self.raw
}

func (self *Data) Raw(data []byte) {
	self.raw = data
}

func (self *Data) Render() {
	var (
		multipart bool = false
		msg bytes.Buffer
		marker string
	)

	if len(self.Attachments) != 0 {
		multipart = true
		marker = makeMarker()
	}

	msg.WriteString(`From: ` + encodeRFC2045(self.From) + "\n")
	msg.WriteString(`To: ` + encodeRFC2045(self.To) + "\n")


	// -------------- head ----------------------------------------------------------

	msg.WriteString("Subject: " + encodeRFC2045(self.Subject) + "\n")
	msg.WriteString("MIME-Version: 1.0\n")
	msg.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\n")
	if multipart {
		msg.WriteString("Content-Type: multipart/mixed;\n	boundary=\"" + marker + "\"\n")
	} else {
		msg.WriteString("Content-Transfer-Encoding: base64\nContent-Type: text/html; charset=\"utf-8\"\n")
	}
	msg.WriteString(strings.Join(self.ExtraHeader, "\n") + "\n")
	// ------------- /head ---------------------------------------------------------

	// ------------- body ----------------------------------------------------------
	if multipart {
		msg.WriteString("--" + marker + "\n")
		msg.WriteString("Content-Transfer-Encoding: base64\nContent-Type: text/html; charset=\"utf-8\"\n\n")
	}
	line76(&msg, base64.StdEncoding.EncodeToString([]byte(self.Html)))
	msg.WriteString("\n")
	// ------------ /body ---------------------------------------------------------

	// ----------- attachments ----------------------------------------------------
	for _, file := range self.Attachments {
		msg.WriteString("\n--" + marker)
		content, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println(err)
		}
		msg.WriteString(fmt.Sprintf("\nContent-Type: %s;\n	name=\"%s\"\nContent-Transfer-Encoding: base64\nContent-Disposition: attachment;\n	filename=\"%s\"\n\n", http.DetectContentType(content), filepath.Base(file), filepath.Base(file)))
		line76(&msg, base64.StdEncoding.EncodeToString(content))
		msg.WriteString("\n")
	}
	// ----------- /attachments ---------------------------------------------------

	self.raw = msg.Bytes()
	return
}

func makeMarker() string {
	b := make([]byte, 30)
	rand.Read(b)
	en := base64.StdEncoding // or URLEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	return string(d)
}


func line76(target *bytes.Buffer, encoded string) {
	nbrLines := len(encoded) / 76
	for i := 0; i < nbrLines; i++ {
		target.WriteString(encoded[i*76:(i+1)*76])
		target.WriteString("\n")
	}
	target.WriteString(encoded[nbrLines*76:])
	target.WriteString("\n")
}

func encodeRFC2045(s string) string {
	return mime.BEncoding.Encode("utf-8", s)
}

func encodeRFC2047(s string) string {
	return mime.QEncoding.Encode("utf-8", s)
}

// Null memory allocate convert
func BytesToString(b []byte) string {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	stringHeader := reflect.StringHeader{
		Data: sliceHeader.Data,
		Len: sliceHeader.Len,
	}
	return *(*string)(unsafe.Pointer(&stringHeader))
}

func StringToBytes(s string) []byte {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	sliceHeader := reflect.SliceHeader{
		Data: stringHeader.Data,
		Len: stringHeader.Len,
		Cap: stringHeader.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&sliceHeader))
}