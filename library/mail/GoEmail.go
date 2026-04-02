package mail

import "gopkg.in/gomail.v2"

// GoMail 实现 Mailer 接口
type GoMail struct {
	Host     string
	Port     int
	UserName string
	Password string
	SSL      bool
}

func (mail *GoMail) Send(c Config) (err error) {
	m := gomail.NewMessage()
	m.SetHeader("From", c.From) // 发件人，必须是你的 QQ 邮箱
	m.SetHeader("To", c.To...)  // 多个收件人
	m.SetHeader("Subject", c.Subject)
	m.SetBody("text/html", c.Body) // 邮件正文（HTML）

	// 添加附件（可选）
	for _, file := range c.Attachments {
		m.Attach(file)
	}

	// QQ 邮箱 SMTP 配置
	//d := gomail.NewDialer("smtp.qq.com", 465, "1252578541@qq.com", "liusweet521!")
	d := gomail.NewDialer(mail.Host, mail.Port, mail.UserName, mail.Password)
	d.SSL = true // QQ 邮箱推荐用 SSL（465端口）

	return d.DialAndSend(m)
}
