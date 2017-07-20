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
	"mime/quotedprintable"
	"reflect"
	"unsafe"
)

func (self *Email) get() []byte {
	return self.raw
}

func (self *Email) SetRawMessage(data []byte) {
	self.raw = data
}

func (self *Email) GetRawMessage() []byte{
	return self.raw
}

func (self *Email) Header(header string) {
	self.headers = append(self.headers, header)
}

func (self *Email) Part(contentType, content string) (err error) {
	var part bytes.Buffer
	if contentType == TypeTextPlain {
		_, err = part.WriteString( "Content-Type: " + contentType + ";\n\t charset=\"utf-8\"\nContent-Transfer-Encoding: quoted-printable\n\n")
		if err != nil {
			return err
		}
		w := quotedprintable.NewWriter(&part)
		w.Write([]byte(content))
		w.Close()
	} else {
		_, err = part.WriteString( "Content-Type: " + contentType + ";\n\t charset=\"utf-8\"\nContent-Transfer-Encoding: base64\n\n")
		if err != nil {
			return err
		}
		err = line76(&part, base64.StdEncoding.EncodeToString([]byte(content)))
		if err != nil {
			return err
		}
	}
	self.parts = append(self.parts, part.Bytes())
	return nil
}

func (self *Email) Attachment(filePath string) (err error) {
	var (
		part    bytes.Buffer
	  	content []byte
	)
	content, err = ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	_, err = part.WriteString(fmt.Sprintf("Content-Type: %s;\n\tname=\"%s\"\nContent-Transfer-Encoding: base64\nContent-Disposition: attachment;\n\tfilename=\"%s\"\n\n", http.DetectContentType(content), filepath.Base(filePath), filepath.Base(filePath)))
	if err != nil {
		return err
	}
	err = line76(&part, base64.StdEncoding.EncodeToString(content))
	if err != nil {
		return err
	}
	self.parts = append(self.parts, part.Bytes())
	return nil
}

func (self *Email) Render() (err error) {
	var (
		msg bytes.Buffer
		marker string
	)

	_, err = msg.WriteString("From: ")
	if err != nil {
		return err
	}
	if self.FromName != "" {
		_, err = msg.WriteString( encodeRFC2045(self.FromName) + " ")
		if err != nil {
			return err
		}
	}
	_, err = msg.WriteString("<" + self.FromEmail + ">\n")
	if err != nil {
		return err
	}
	_, err = msg.WriteString("To: ")
	if err != nil {
		return err
	}
	if self.ToName != "" {
		_, err = msg.WriteString(encodeRFC2045(self.ToName) + " ")
		if err != nil {
			return err
		}
	}
	_, err = msg.WriteString("<" + self.ToEmail + ">\n")
	if err != nil {
		return err
	}

	// -------------- head ----------------------------------------------------------
	_, err = msg.WriteString("Subject: " + encodeRFC2045(self.Subject) + "\n")
	if err != nil {
		return err
	}
	_, err = msg.WriteString("MIME-Version: 1.0\n")
	if err != nil {
		return err
	}
	_, err = msg.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\n")
	if err != nil {
		return err
	}


	_, err = msg.WriteString(strings.Join(self.headers, "\n") + "\n")
	if err != nil {
		return err
	}

	if len(self. parts) > 1 {
		marker = makeMarker()
		_, err = msg.WriteString("Content-Type: multipart/mixed;\n\tboundary=\"" + marker + "\"\n\n")
		if err != nil {
			return err
		}
	}

	// ------------- /head ---------------------------------------------------------

	// ------------- body ----------------------------------------------------------
	for i := range self. parts {
		if len(self.parts) > 1 {
			_, err = msg.WriteString("--" + marker + "\n")
		}
		_, err = msg.Write(self.parts[i])
		if err != nil {
			return err
		}
		_, err = msg.WriteString("\n")
		if err != nil {
			return err
		}
	}
	// ------------- /body ----------------------------------------------------------

	self.raw = msg.Bytes()
	return nil
}

func makeMarker() string {
	b := make([]byte, 30)
	rand.Read(b)
	en := base64.StdEncoding // or URLEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	return string(d)
}


func line76(target *bytes.Buffer, encoded string) (err error) {
	nbrLines := len(encoded) / 76
	for i := 0; i < nbrLines; i++ {
		_, err = target.WriteString(encoded[i*76:(i+1)*76])
		if err != nil {
			return err
		}
		_, err = target.WriteString("\n")
		if err != nil {
			return err
		}
	}
	_, err = target.WriteString(encoded[nbrLines*76:])
	if err != nil {
		return err
	}
	_, err = target.WriteString("\n")
	if err != nil {
		return err
	}

	return nil
}

func encodeRFC2045(s string) string {
	return mime.BEncoding.Encode("utf-8", s)
}

//
//func encodeRFC2047(s string) string {
//	return mime.QEncoding.Encode("utf-8", s)
//}

// Null memory allocate convert
func BytesToString(b []byte) string {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	stringHeader := reflect.StringHeader{
		Data: sliceHeader.Data,
		Len: sliceHeader.Len,
	}
	return *(*string)(unsafe.Pointer(&stringHeader))
}

//func StringToBytes(s string) []byte {
//	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
//	sliceHeader := reflect.SliceHeader{
//		Data: stringHeader.Data,
//		Len: stringHeader.Len,
//		Cap: stringHeader.Len,
//	}
//	return *(*[]byte)(unsafe.Pointer(&sliceHeader))
//}