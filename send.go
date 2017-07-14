package directEmail

import (
	"net"
	"fmt"
	"strings"
	"time"
	"net/smtp"
	"errors"
	"golang.org/x/net/idna"
)

type Email struct {
	Ip    net.Addr
	MapIp map[string]string
	From  string
	To    string
	Data  Data
}

func New() Email {
	return Email{Ip: &net.TCPAddr{
		IP:  net.ParseIP("127.0.0.1"),
	}}
}

func (self *Email) Send() error {

	var myGlobalIP string
	myIp,_, err := net.SplitHostPort(self.Ip.String())
	myGlobalIP, ok := self.MapIp[myIp]
	if !ok {
		myGlobalIP = myIp
	}

	name, err := net.LookupAddr(myGlobalIP)
	if err != nil && len(name) < 1 {
		return err
	}

	splitEmail := strings.SplitN(self.To, "@", 2)
	if len(splitEmail) != 2 {
		return errors.New("550 Bad email")
	}

	domain, err := idna.ToASCII(splitEmail[1])
	if err != nil {
		return errors.New(fmt.Sprintf("550 Domain name failed: %v", err))
	}

	addr := &net.TCPAddr{
		IP: net.ParseIP(self.Ip.String()),
	}
	iface := net.Dialer{LocalAddr: addr}

	record, err := net.LookupMX(domain)
	if err != nil {
		return errors.New(fmt.Sprintf("550 %v", err))
	}

	var (
		conn net.Conn
		server string
	)
	for i := range record {
		server = strings.TrimRight(strings.TrimSpace(record[i].Host), ".")
		conn, err = iface.Dial("tcp", net.JoinHostPort(server, "25"))
		if err == nil {
			break
		}
	}
	if err != nil {
		return errors.New(fmt.Sprintf("550 %v", err))
	}
	conn.SetDeadline(time.Now().Add(5 * time.Minute))

	c, err := smtp.NewClient(conn, server)
	if err != nil {
		return err
	}

	if err := c.Hello(strings.TrimRight(name[0], ".")); err != nil {
		return err
	}

	if err := c.Mail(self.From); err != nil {
		return err
	}

	if err := c.Rcpt(self.To); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, string(self.Data.raw))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()

}
