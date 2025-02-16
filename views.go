package main

import (
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var css = []string{"bundle.css", "remixicon.css"}
var cssHeader string

func (app *appContext) loadCSSHeader() string {
	l := len(css)
	h := ""
	for i, f := range css {
		h += "<" + app.URLBase + "/css/" + f + ">; rel=preload; as=style"
		if l > 1 && i != (l-1) {
			h += ", "
		}
	}
	return h
}

func (app *appContext) getURLBase(gc *gin.Context) string {
	if strings.HasPrefix(gc.Request.URL.String(), app.URLBase) {
		return app.URLBase
	}
	return ""
}

func gcHTML(gc *gin.Context, code int, file string, templ gin.H) {
	gc.Header("Cache-Control", "no-cache")
	gc.HTML(code, file, templ)
}

func (app *appContext) pushResources(gc *gin.Context, admin bool) {
	if pusher := gc.Writer.Pusher(); pusher != nil {
		app.debug.Println("Using HTTP2 Server push")
		if admin {
			toPush := []string{"/js/admin.js", "/js/theme.js", "/js/lang.js", "/js/modal.js", "/js/tabs.js", "/js/invites.js", "/js/accounts.js", "/js/settings.js", "/js/profiles.js", "/js/common.js"}
			for _, f := range toPush {
				if err := pusher.Push(app.URLBase+f, nil); err != nil {
					app.debug.Printf("Failed HTTP2 ServerPush of \"%s\": %+v", f, err)
				}
			}
		}
	}
	gc.Header("Link", cssHeader)
}

func (app *appContext) getLang(gc *gin.Context, chosen string) string {
	lang := gc.Query("lang")
	if lang == "" {
		lang = chosen
	} else if _, ok := app.storage.lang.Admin[lang]; !ok {
		lang = chosen
	}
	return lang
}

func (app *appContext) AdminPage(gc *gin.Context) {
	app.pushResources(gc, true)
	lang := app.getLang(gc, app.storage.lang.chosenAdminLang)
	emailEnabled, _ := app.config.Section("invite_emails").Key("enabled").Bool()
	notificationsEnabled, _ := app.config.Section("notifications").Key("enabled").Bool()
	ombiEnabled := app.config.Section("ombi").Key("enabled").MustBool(false)
	var license string
	l, err := fs.ReadFile(localFS, "LICENSE")
	if err != nil {
		app.debug.Printf("Failed to load LICENSE: %s", err)
		license = ""
	}
	license = string(l)
	gcHTML(gc, http.StatusOK, "admin.html", gin.H{
		"urlBase":         app.getURLBase(gc),
		"cssClass":        app.cssClass,
		"contactMessage":  "",
		"email_enabled":   emailEnabled,
		"notifications":   notificationsEnabled,
		"version":         version,
		"commit":          commit,
		"ombiEnabled":     ombiEnabled,
		"username":        !app.config.Section("email").Key("no_username").MustBool(false),
		"strings":         app.storage.lang.Admin[lang].Strings,
		"quantityStrings": app.storage.lang.Admin[lang].QuantityStrings,
		"language":        app.storage.lang.Admin[lang].JSON,
		"langName":        lang,
		"license":         license,
	})
}

func (app *appContext) ResetPassword(gc *gin.Context) {
	pin := gc.Query("pin")
	if pin == "" {
		app.NoRouteHandler(gc)
		return
	}
	app.pushResources(gc, false)
	lang := app.getLang(gc, app.storage.lang.chosenPWRLang)
	data := gin.H{
		"urlBase":        app.getURLBase(gc),
		"contactMessage": app.config.Section("ui").Key("contact_message").String(),
		"strings":        app.storage.lang.PasswordReset[lang].Strings,
		"success":        false,
	}
	resp, status, err := app.jf.ResetPassword(pin)
	if status == 200 && err == nil && resp.Success {
		data["success"] = true
		data["pin"] = pin
	} else {
		app.err.Printf("Password Reset failed (%d): %v", status, err)
	}
	gcHTML(gc, http.StatusOK, "password-reset.html", data)
}

func (app *appContext) InviteProxy(gc *gin.Context) {
	app.pushResources(gc, false)
	code := gc.Param("invCode")
	lang := app.getLang(gc, app.storage.lang.chosenFormLang)
	/* Don't actually check if the invite is valid, just if it exists, just so the page loads quicker. Invite is actually checked on submit anyway. */
	// if app.checkInvite(code, false, "") {
	inv, ok := app.storage.invites[code]
	if !ok {
		gcHTML(gc, 404, "invalidCode.html", gin.H{
			"cssClass":       app.cssClass,
			"contactMessage": app.config.Section("ui").Key("contact_message").String(),
		})
		return
	}
	if key := gc.Query("key"); key != "" && app.config.Section("email_confirmation").Key("enabled").MustBool(false) {
		validKey := false
		keyIndex := -1
		for i, k := range inv.Keys {
			if k == key {
				validKey = true
				keyIndex = i
				break
			}
		}
		fail := func() {
			gcHTML(gc, 404, "404.html", gin.H{
				"cssClass":       app.cssClass,
				"contactMessage": app.config.Section("ui").Key("contact_message").String(),
			})
		}
		if !validKey {
			fail()
			return
		}
		token, err := jwt.Parse(key, checkToken)
		if err != nil {
			fail()
			app.err.Printf("Failed to parse key: %s", err)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		expiryUnix, err := strconv.ParseInt(claims["exp"].(string), 10, 64)
		if err != nil {
			fail()
			app.err.Printf("Failed to parse key expiry: %s", err)
			return
		}
		expiry := time.Unix(expiryUnix, 0)
		if !(ok && token.Valid && claims["invite"].(string) == code && claims["type"].(string) == "confirmation" && expiry.After(time.Now())) {
			fail()
			app.debug.Printf("Invalid key")
			return
		}
		req := newUserDTO{
			Email:    claims["email"].(string),
			Username: claims["username"].(string),
			Password: claims["password"].(string),
			Code:     claims["invite"].(string),
		}
		_, success := app.newUser(req, true)
		if !success {
			fail()
			return
		}
		gcHTML(gc, http.StatusOK, "create-success.html", gin.H{
			"strings":        app.storage.lang.Form[lang].Strings,
			"successMessage": app.config.Section("ui").Key("success_message").String(),
			"contactMessage": app.config.Section("ui").Key("contact_message").String(),
			"jfLink":         app.config.Section("jellyfin").Key("public_server").String(),
		})
		inv, ok := app.storage.invites[code]
		if ok {
			l := len(inv.Keys)
			inv.Keys[l-1], inv.Keys[keyIndex] = inv.Keys[keyIndex], inv.Keys[l-1]
			app.storage.invites[code] = inv
		}
		return
	}
	email := app.storage.invites[code].Email
	if strings.Contains(email, "Failed") {
		email = ""
	}
	gcHTML(gc, http.StatusOK, "form-loader.html", gin.H{
		"urlBase":           app.getURLBase(gc),
		"cssClass":          app.cssClass,
		"contactMessage":    app.config.Section("ui").Key("contact_message").String(),
		"helpMessage":       app.config.Section("ui").Key("help_message").String(),
		"successMessage":    app.config.Section("ui").Key("success_message").String(),
		"jfLink":            app.config.Section("jellyfin").Key("public_server").String(),
		"validate":          app.config.Section("password_validation").Key("enabled").MustBool(false),
		"requirements":      app.validator.getCriteria(),
		"email":             email,
		"username":          !app.config.Section("email").Key("no_username").MustBool(false),
		"strings":           app.storage.lang.Form[lang].Strings,
		"validationStrings": app.storage.lang.Form[lang].validationStringsJSON,
		"notifications":     app.storage.lang.Form[lang].notificationsJSON,
		"code":              code,
		"confirmation":      app.config.Section("email_confirmation").Key("enabled").MustBool(false),
		"userExpiry":        inv.UserExpiry,
		"userExpiryDays":    inv.UserDays,
		"userExpiryHours":   inv.UserHours,
		"userExpiryMinutes": inv.UserMinutes,
		"userExpiryMessage": app.storage.lang.Form[lang].Strings.get("yourAccountIsValidUntil"),
	})
}

func (app *appContext) NoRouteHandler(gc *gin.Context) {
	app.pushResources(gc, false)
	gcHTML(gc, 404, "404.html", gin.H{
		"cssClass":       app.cssClass,
		"contactMessage": app.config.Section("ui").Key("contact_message").String(),
	})
}
