package email

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
)

type Email struct {
	FromName, FromEmail, ToName, ToEmail, Subject string
	Message                                       string
}

type Client struct {
	Email, Password string
	Host, Port      string
}

func (client *Client) Send(em *Email) error {
	from := mail.Address{em.FromName, em.FromEmail}
	to := mail.Address{em.ToName, em.ToEmail}

	auth := smtp.PlainAuth("", client.Email, client.Password, client.Host)

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = em.Subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(em.Message))

	messageBytes := []byte(message)

	err := smtp.SendMail(client.Host+":"+client.Port, auth, from.Address, []string{to.Address}, messageBytes)
	return err
}
