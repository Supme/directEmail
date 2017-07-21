package directEmail

import (
	"reflect"
	"unsafe"
	"encoding/base64"
	"bytes"
	"math/rand"
	"mime"
)


func makeMarker() string {
	b := make([]byte, 30)
	rand.Read(b)
	en := base64.StdEncoding // or URLEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	return "_" + string(d) + "_"
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
//func BytesToString(b []byte) string {
//	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
//	stringHeader := reflect.StringHeader{
//		Data: sliceHeader.Data,
//		Len: sliceHeader.Len,
//	}
//	return *(*string)(unsafe.Pointer(&stringHeader))
//}

//func StringToBytes(s string) []byte {
//	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
//	sliceHeader := reflect.SliceHeader{
//		Data: stringHeader.Data,
//		Len: stringHeader.Len,
//		Cap: stringHeader.Len,
//	}
//	return *(*[]byte)(unsafe.Pointer(&sliceHeader))
//}