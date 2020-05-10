package emails

import (
	"fmt"
	"strconv"

	"github.com/sendgrid/rest"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/viper"

	"github.com/bryanmorgan/time-tracking-api/logger"
)

func SendForgotPasswordEmail(name string, email string, resetUrl string) error {
	m := mail.NewV3Mail()
	fromEmailName := viper.GetString("email.fromName")
	fromEmailAddress := viper.GetString("email.fromAddress")

	e := mail.NewEmail(fromEmailName, fromEmailAddress)
	m.SetFrom(e)

	m.Subject = "Reset Your Password"
	p := mail.NewPersonalization()

	testMode := viper.GetBool("email.testMode")
	if testMode {
		logger.Log.Warn("|Email Test Mode| : " + email)
		p.AddTos(mail.NewEmail(name, viper.GetString("email.testToEmail")))
	} else {
		p.AddTos(mail.NewEmail(name, email))
	}

	plainTextContent := "Dear %name%, please go to %resetUrl% to reset your password"
	c := mail.NewContent("text/plain", plainTextContent)
	m.AddContent(c)

	htmlContent := `
	Dear %name%,<br/><br/>
	We received your request to <strong>Reset Your Password</strong><br/><br/>
	Click here to <a href="%resetUrl%">reset your password</a><br/><br/><br/>
	If you did not request to reset your password please ignore this request or contact us.<br/><br/><br/>
	Regards,<br/>
	%emailSignature%
	`
	c = mail.NewContent("text/html", htmlContent)
	m.AddContent(c)

	//m.SetTemplateID("13b8f94f-bcae-4ec6-b752-70d6cb59f932")
	emailSignatureName := viper.GetString("email.emailSignatureName")
	p.SetHeader("X-Test", "test")
	p.SetSubstitution("%name%", name)
	p.SetSubstitution("%resetUrl%", resetUrl)
	p.SetSubstitution("%emailSignature%", emailSignatureName)
	m.AddCategories("Forgot Password")
	m.AddPersonalizations(p)

	response, err := sendEmail(m)
	if err != nil {
		logger.Log.Error("Reset password email failed: " + err.Error())
		return err
	}
	if response != nil {
		logger.Log.Info("Sent reset password email to " + email + " [" + strconv.Itoa(response.StatusCode) + "]: " + response.Body)
	}

	return nil
}

func sendEmail(m *mail.SGMailV3) (*rest.Response, error) {
	var response *rest.Response
	var err error

	if viper.GetBool("email.enabled") {
		request := sendgrid.GetRequest(viper.GetString("email.sendGrid.apiKey"), "/v3/mail/send", "https://api.sendgrid.com")
		request.Method = "POST"
		request.Body = mail.GetRequestBody(m)
		response, err = sendgrid.API(request)
		if err != nil {
			return nil, err
		}
	} else {
		if len(m.Personalizations) > 0 && len(m.Personalizations[0].To) > 0 {
			email := m.Personalizations[0].To[0]
			logger.Log.Warn("Email sending disabled. Email not sent: " + fmt.Sprintf("%s [%s]", email.Address, email.Name))
		}
	}

	return response, nil
}
