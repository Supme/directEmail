# directEmail

Example

```
package main

import (
	"github.com/supme/directEmail"
	"time"
	"fmt"
)

func main() {
	email := directEmail.New()

	email.FromEmail = "sender@example.com"
	email.FromName = "Отправитель"

	email.ToEmail = "user@example.com"
	email.ToName = "Получатель"

	email.Header(fmt.Sprintf("Message-ID: <test_message_%d>", time.Now().Unix()))
	email.Header("Content-Language: ru")
	email.Subject = "Тест отправки email"

	email.Part(directEmail.TypeTextHTML, `
<h2>My email</h2>
<p>Текст моего сообщения</p>
	`)

	email.Part(directEmail.TypeTextPlain, `
My email
Текст моего сообщения
`)

  email.Attachment("/path/to/attach/file.jpg")

	email.Render()
	print("\n", string(email.GetRawMessage()), "\n")

	err := email.Send()
	if err != nil {
		print("Send email with error:", err.Error())
	}

}
```
