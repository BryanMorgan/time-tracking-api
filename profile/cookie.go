package profile

import (
	"net/http"
	"time"

	"github.com/spf13/viper"
)


func AddSessionCookie(w http.ResponseWriter, token string, domain string) {
	expiration := GetSessionExpiration()
	cookie := http.Cookie{
		Name:     GetSessionCookieName(),
		Value:    token,
		Path:     "/",
		Expires:  expiration,
		Domain:   domain,
		HttpOnly: true,
		Secure:   viper.GetBool("session.secureCookie"),
	}

	// TODO: Verify SameSite=Strict|Lax attribute
	cookieString := cookie.String() + "; SameSite=lax"
	w.Header().Set("Set-Cookie", cookieString)
}

func ClearSessionCookie(w http.ResponseWriter, domain string) {
	cookie := http.Cookie{
		Name:     GetSessionCookieName(),
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		Domain:   domain,
		HttpOnly: true,
		Secure:   viper.GetBool("session.secureCookie"),
	}
	http.SetCookie(w, &cookie)
}

func GetSessionCookieFromRequest(r *http.Request) *http.Cookie {
	cookie, _ := r.Cookie(GetSessionCookieName())
	return cookie
}

func GetSessionExpiration() time.Time {
	return time.Now().Add(time.Minute * time.Duration(GetCookieExpirationMinutes()))
}

func GetCookieExpirationMinutes() int {
	return viper.GetInt("session.cookieExpirationMinutes")
}
func GetSessionCookieName() string {
	return viper.GetString("session.cookieName")
}
