package mail

// Config 邮件发送配置
type Config struct {
	From        string
	To          []string
	Subject     string
	Body        string
	Attachments []string
}

// Mailer 邮件发送接口
type Mailer interface {
	Send(c Config) error
}

type MailService struct {
	Mailer Mailer
}

func (s *MailService) Send(c Config) error {
	return s.Mailer.Send(c)
}
