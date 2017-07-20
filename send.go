package directEmail

import (
	"net"
	"fmt"
	"strings"
	"time"
	"net/smtp"
	"errors"
	"golang.org/x/net/idna"
	"golang.org/x/net/proxy"
)

const (
	TypeTextHTML = "text/html"
	TypeTextPlain = "text/plain"
)

type Email struct {
	Ip    	  string
	Host      string
	MapIp map[string]string

	FromEmail string
	FromName  string
	ToEmail   string
	ToName    string
	Subject   string

	headers   []string
	parts     [][]byte
	raw       []byte
}

func New() Email {
	return Email{}
}

func (self *Email) Send() error {
	var err error

	splitEmail := strings.SplitN(self.ToEmail, "@", 2)
	if len(splitEmail) != 2 {
		return errors.New("550 Bad email")
	}

	domain, err := idna.ToASCII(splitEmail[1])
	if err != nil {
		return errors.New(fmt.Sprintf("550 Domain name failed: %v", err))
	}

	client, err := self.dial(domain)
	if err != nil {
		return errors.New(fmt.Sprintf("550 %v", err))
	}

	if err := client.Hello(strings.TrimRight(self.Host, ".")); err != nil {
		return err
	}

	if err := client.Mail(self.FromEmail); err != nil {
		return err
	}

	if err := client.Rcpt(self.ToEmail); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, BytesToString(self.raw))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()

}

type dialer func(network, address string) (net.Conn, error)

func (self *Email) dial(domain string) (client *smtp.Client, err error) {
	var (
		conn     net.Conn
		dialFunc dialer
	)

	if self.Ip == "" {
		iface := net.Dialer{}
		dialFunc = iface.Dial
debug("Dial function is default network interface\n")
	} else {
		if strings.ToLower(self.Ip[0:8]) == "socks://" {
			iface, err := proxy.SOCKS5("tcp", self.Ip[8:], nil, proxy.FromEnvironment())
			if err != nil {
				return nil, err
			}
			dialFunc = iface.Dial
debug("Dial function is socks proxy from ", self.Ip[8:] ,"\n")
		} else {
			addr := &net.TCPAddr{
				IP: net.ParseIP(self.Ip),
			}
			iface := net.Dialer{LocalAddr: addr}
			dialFunc = iface.Dial
debug("Dial function is ", addr.String() ," network interface\n")
		}
	}

	records, err := net.LookupMX(domain)
	if err != nil {
		return
	}
debug("MX for domain:\n")
for i:=range records {
debug(" - ", records[i].Pref, " ", records[i].Host, "\n")
}

	for i := range records {
		server := strings.TrimRight(strings.TrimSpace(records[i].Host), ".")
debug("Connect to server ", server, "\n")
		conn, err = dialFunc("tcp", net.JoinHostPort(server, "25"))
		if err != nil {
debug("Not connected\n")
			continue
		}
debug("Connected\n")
		client, err = smtp.NewClient(conn, server)
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}

	conn.SetDeadline(time.Now().Add(5 * time.Minute)) // SMTP RFC

	if self.Ip == "" {
		self.Ip = conn.LocalAddr().String()
	}

	if self.Host == "" {
		var myGlobalIP string
		myIp,_, err := net.SplitHostPort(strings.TrimLeft(self.Ip, "socks://"))
		myGlobalIP, ok := self.MapIp[myIp]
		if !ok {
			myGlobalIP = myIp
		}
		names, err := net.LookupAddr(myGlobalIP)
		if err != nil && len(names) < 1 {
			return nil, err
		}
debug("LookUp ", myGlobalIP, " this result ", names[0], "\n")
		self.Host = names[0]
	}

	return
}

func debug(args ...interface{}) {
	fmt.Print(args...)
}