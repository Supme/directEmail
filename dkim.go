package directEmail

import (
	dkim "github.com/toorop/go-dkim"
)

// DkimSign add DKIM signature email
// Generate private key:
//  openssl genrsa -out /path/to/key/example.com.key 2048
// Generate public key:
//  openssl rsa -in /path/to/key/example.com.key -pubout
// Add public key to DNS myselector._domainkey.example.com TXT record
//  k=rsa; p=MIGfMA0GC...
func (self *Email) DkimSign(selector string, privateKey []byte) error {
	domain, err := self.domainFromEmail(self.FromEmail)
	if err != nil {
		return err
	}
	options := dkim.NewSigOptions()
	options.PrivateKey = privateKey
	options.Domain = domain
	options.Selector = selector
	options.Algo = "rsa-sha1"
	options.Headers = []string{"from", "to", "subject"}
	options.AddSignatureTimestamp = true
	options.Canonicalization = "simple/simple"

	email :=  self.GetRawMessageBytes()

	if self.bodyLenght >= 50 {
		options.BodyLength = 50
	}

	err = dkim.Sign(&email, options)
	if err != nil {
		return err
	}
	self.SetRawMessageBytes(email)

	return nil
}
