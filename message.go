package directEmail

import (
	"time"
	"fmt"
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"strings"
	"path/filepath"
	"mime/quotedprintable"
)

func (self *Email) SetRawMessageBytes(data []byte) error {
	_, err := self.raw.Write(data)
	return err
}

func (self *Email) SetRawMessageString(data string) error {
	_, err := self.raw.WriteString(data)
	return err
}

func (self *Email) GetRawMessageBytes() []byte {
	return self.raw.Bytes()
}

func (self *Email) GetRawMessageString() string {
	return self.raw.String()
}


func (self *Email) Header(headers ...string) {
	for i := range headers {
		self.headers = append(self.headers, headers[i])
	}
}

func (self *Email) TextPlain(content string) (err error) {
	var part bytes.Buffer
	defer part.Reset()
	_, err = part.WriteString( "Content-Type: text/plain;\n\t charset=\"utf-8\"\nContent-Transfer-Encoding: quoted-printable\n\n")
	if err != nil {
		return err
	}
	w := quotedprintable.NewWriter(&part)
	w.Write([]byte(strings.TrimLeft(content, "\n")))
	w.Close()
	self.textPlain = part.Bytes()
	return nil
}

func (self *Email) TextHtml(content string) (err error) {
	var part bytes.Buffer
	defer part.Reset()
	_, err = part.WriteString( "Content-Type: text/html;\n\t charset=\"utf-8\"\nContent-Transfer-Encoding: base64\n\n")
	if err != nil {
		return err
	}
	err = line76(&part, base64.StdEncoding.EncodeToString([]byte(content)))
	if err != nil {
		return err
	}
	self.textHtml = part.Bytes()
	return nil
}

func (self *Email) TextHtmlWithRelated(content string, files ...string) (err error) {
	var (
		part bytes.Buffer
		data []byte
	)
	defer part.Reset()

	marker := makeMarker()
	_, err = part.WriteString("Content-Type: multipart/related;\n\tboundary=\"" + marker + "\"\n")
	if err != nil {
		return err
	}

	_, err = part.WriteString("\n--" + marker + "\n")
	if err != nil {
		return err
	}
	_, err = part.WriteString( "Content-Type: text/html;\n\t charset=\"utf-8\"\nContent-Transfer-Encoding: base64\n\n")
	if err != nil {
		return err
	}
	err = line76(&part, base64.StdEncoding.EncodeToString([]byte(content)))
	if err != nil {
		return err
	}

	for i := range files {
		_, err = part.WriteString("\n--" + marker + "\n")
		if err != nil {
			return err
		}

		data, err = ioutil.ReadFile(files[i])
		if err != nil {
			return err
		}
		_, err = part.WriteString(fmt.Sprintf("Content-Type: %s;\n\tname=\"%s\"\nContent-Transfer-Encoding: base64\nContent-ID: <%s>\nContent-Disposition: inline;\n\tfilename=\"%s\"; size=%d;\n\n", http.DetectContentType(data), filepath.Base(files[i]), filepath.Base(files[i]), filepath.Base(files[i]), len(data)))
		if err != nil {
			return err
		}
		err = line76(&part, base64.StdEncoding.EncodeToString(data))
		if err != nil {
			return err
		}

	}
	_, err = part.WriteString("\n--" + marker + "--\n")

	self.textHtml = part.Bytes()
	return nil
}

func (self *Email) Attachment(file string) (err error) {
	var (
		part    bytes.Buffer
	  	content []byte
	)
	defer part.Reset()
	content, err = ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	_, err = part.WriteString(fmt.Sprintf("Content-Type: %s;\n\tname=\"%s\"\nContent-Transfer-Encoding: base64\nContent-Disposition: attachment;\n\tfilename=\"%s\"; size=%d;\n\n", http.DetectContentType(content), filepath.Base(file), filepath.Base(file), len(content)))
	if err != nil {
		return err
	}
	err = line76(&part, base64.StdEncoding.EncodeToString(content))
	if err != nil {
		return err
	}
	self.attachments = append(self.attachments, part.Bytes())
	return nil
}

func (self *Email) Render() (err error) {
	var (
		marker  string
	)

	// -------------- head ----------------------------------------------------------
	_, err = self.raw.WriteString("From: ")
	if err != nil {
		return err
	}
	if self.FromName != "" {
		_, err = self.raw.WriteString( encodeRFC2045(self.FromName) + " ")
		if err != nil {
			return err
		}
	}
	_, err = self.raw.WriteString("<" + self.FromEmail + ">\n")
	if err != nil {
		return err
	}
	_, err = self.raw.WriteString("To: ")
	if err != nil {
		return err
	}
	if self.ToName != "" {
		_, err = self.raw.WriteString(encodeRFC2045(self.ToName) + " ")
		if err != nil {
			return err
		}
	}
	_, err = self.raw.WriteString("<" + self.ToEmail + ">\n")
	if err != nil {
		return err
	}

	_, err = self.raw.WriteString("Subject: " + encodeRFC2045(self.Subject) + "\n")
	if err != nil {
		return err
	}
	_, err = self.raw.WriteString("MIME-Version: 1.0\n")
	if err != nil {
		return err
	}
	_, err = self.raw.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\n")
	if err != nil {
		return err
	}

	_, err = self.raw.WriteString(strings.Join(self.headers, "\n") + "\n")
	if err != nil {
		return err
	}

	// Email has text and attachment?
	if len(self.attachments) > 0 {
		marker = makeMarker()
		_, err = self.raw.WriteString("Content-Type: multipart/mixed;\n\tboundary=\"" + marker + "\"\n")
		if err != nil {
			return err
		}
	}

	// ------------- /head ---------------------------------------------------------

	if len(self.textPlain) > 0 || len(self.textHtml) > 0 {
		if marker != "" {
			_, err = self.raw.WriteString("\n--" + marker + "\n")
		}
		err = self.renderText()
		if err != nil {
			return err
		}
	}

	// ------------- attachments ----------------------------------------------------------
	for i := range self.attachments {
		_, err = self.raw.WriteString("\n--" + marker + "\n")
		if err != nil {
			return err
		}
		_, err = self.raw.Write(self.attachments[i])
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
	}
	// ------------- /attachments ----------------------------------------------------------

	if marker != "" {
		_, err = self.raw.WriteString("\n--" + marker + "--\n")
	}

	return nil
}

func (self *Email) renderText() error {
	var (
		marker string
		err error
	)
	if len(self.textPlain) > 0 && len(self.textHtml) > 0 {
		marker = makeMarker()
		_, err = self.raw.WriteString("Content-Type: multipart/alternative;\n\tboundary=\"" + marker + "\"\n")
		if err != nil {
			return err
		}
	}

	if marker != "" {
			_, err = self.raw.WriteString("\n--" + marker + "\n")
	}

	if len(self.textPlain) > 0 {
		_, err = self.raw.Write(self.textPlain)
		if err != nil {
			return err
		}
	}

	if marker != "" {
		_, err = self.raw.WriteString("\n--" + marker + "\n")
	}

	if len(self.textHtml) > 0 {
		_, err = self.raw.Write(self.textHtml)
		if err != nil {
			return err
		}
	}

	if marker != "" {
		_, err = self.raw.WriteString("\n--" + marker + "--\n")
	}

	return nil
}
