package service

import (
	"github.com/be-ys-cloud/dory-server/internal/authentication/token"
	"github.com/be-ys-cloud/dory-server/internal/configuration"
	"github.com/be-ys-cloud/dory-server/internal/ldap"
	"github.com/be-ys-cloud/dory-server/internal/mailer"
	"github.com/be-ys-cloud/dory-server/internal/structures"
	"github.com/sirupsen/logrus"
	"net/url"
)

func AskMail(user structures.UserAsk, kind string) error {
	// Get user email
	email, err := ldap.GetUserMail(user.Username)

	if err != nil {
		logrus.Warnf("Could not get email for user %s! Unable to send link, aborting request.", user.Username)
		return &structures.CustomError{Text: "could not find your email on active directory server", HttpCode: 400}
	}

	//Create new unique key
	key := token.CreateKey(user.Username)

	//Send mail to user
	err = mailer.SendMail(kind, email, struct {
		Name     string
		URL      string
		Username string
		Token    string
		LDAP     string
	}{
		Name:     user.Username,
		URL:      configuration.Configuration.FrontAddress,
		Username: url.QueryEscape(user.Username),
		Token:    url.QueryEscape(key),
		LDAP:     configuration.Configuration.LDAPServer.Address,
	})

	if err != nil {
		logrus.Warnf("Could not send to user %s! Unable to send link, aborting request.", user.Username)
		token.DeleteKey(user.Username)
		return &structures.CustomError{Text: "email gateway is not reachable, try again later", HttpCode: 503}
	}

	if configuration.Configuration.Features.EnableAudit {
		logrus.WithField("user", user.Username).Infof("[AUDIT] Sent mail to user for kind (%s)", kind)
	}

	return nil
}
