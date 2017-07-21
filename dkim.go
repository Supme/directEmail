package directEmail

import (
	dkim "github.com/toorop/go-dkim"
	"strings"
	"fmt"
	"errors"
	"golang.org/x/net/idna"
)

// DkimSign add DKIM signature email
func (self *Email) DkimSign(selector string, privateKey []byte) error {
	splitEmail := strings.SplitN(self.FromEmail, "@", 2)
	if len(splitEmail) != 2 {
		return errors.New("Bad from email address")
	}

	domain, err := idna.ToASCII(strings.TrimRight(splitEmail[1], "."))
	if err != nil {
		return errors.New(fmt.Sprintf("Domain name failed: %v", err))
	}
	options := dkim.NewSigOptions()
	options.PrivateKey = privateKey
	options.Domain = domain
	options.Selector = selector
	options.Headers = []string{"from", "to", "date", "subject"}
	options.AddSignatureTimestamp = true
	options.Canonicalization = "relaxed/relaxed"

	email :=  self.GetRawMessageBytes()
	err = dkim.Sign(&email, options)
	if err != nil {
		return err
	}
	self.SetRawMessageBytes(email)

	return nil
}
