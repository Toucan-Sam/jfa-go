package main

type langMeta struct {
	Name string `json:"name"`
}

type quantityString struct {
	Singular string `json:"singular"`
	Plural   string `json:"plural"`
}

type adminLangs map[string]adminLang

func (ls *adminLangs) getOptions() [][2]string {
	opts := make([][2]string, len(*ls))
	i := 0
	for key, lang := range *ls {
		opts[i] = [2]string{key, lang.Meta.Name}
		i++
	}
	return opts
}

type commonLangs map[string]commonLang

type commonLang struct {
	Meta    langMeta    `json:"meta"`
	Strings langSection `json:"strings"`
}

type adminLang struct {
	Meta            langMeta                  `json:"meta"`
	Strings         langSection               `json:"strings"`
	Notifications   langSection               `json:"notifications"`
	QuantityStrings map[string]quantityString `json:"quantityStrings"`
	JSON            string
}

type formLangs map[string]formLang

func (ls *formLangs) getOptions() [][2]string {
	opts := make([][2]string, len(*ls))
	i := 0
	for key, lang := range *ls {
		opts[i] = [2]string{key, lang.Meta.Name}
		i++
	}
	return opts
}

type formLang struct {
	Meta                  langMeta    `json:"meta"`
	Strings               langSection `json:"strings"`
	Notifications         langSection `json:"notifications"`
	notificationsJSON     string
	ValidationStrings     map[string]quantityString `json:"validationStrings"`
	validationStringsJSON string
}

type pwrLangs map[string]pwrLang

func (ls *pwrLangs) getOptions() [][2]string {
	opts := make([][2]string, len(*ls))
	i := 0
	for key, lang := range *ls {
		opts[i] = [2]string{key, lang.Meta.Name}
		i++
	}
	return opts
}

type pwrLang struct {
	Meta    langMeta    `json:"meta"`
	Strings langSection `json:"strings"`
}

type emailLangs map[string]emailLang

func (ls *emailLangs) getOptions() [][2]string {
	opts := make([][2]string, len(*ls))
	i := 0
	for key, lang := range *ls {
		opts[i] = [2]string{key, lang.Meta.Name}
		i++
	}
	return opts
}

type emailLang struct {
	Meta              langMeta    `json:"meta"`
	Strings           langSection `json:"strings"`
	UserCreated       langSection `json:"userCreated"`
	InviteExpiry      langSection `json:"inviteExpiry"`
	PasswordReset     langSection `json:"passwordReset"`
	UserDeleted       langSection `json:"userDeleted"`
	InviteEmail       langSection `json:"inviteEmail"`
	WelcomeEmail      langSection `json:"welcomeEmail"`
	EmailConfirmation langSection `json:"emailConfirmation"`
	UserExpired       langSection `json:"userExpired"`
}

type setupLangs map[string]setupLang

type setupLang struct {
	Meta               langMeta    `json:"meta"`
	Strings            langSection `json:"strings"`
	StartPage          langSection `json:"startPage"`
	EndPage            langSection `json:"endPage"`
	General            langSection `json:"general"`
	Updates            langSection `json:"updates"`
	Language           langSection `json:"language"`
	Login              langSection `json:"login"`
	JellyfinEmby       langSection `json:"jellyfinEmby"`
	Ombi               langSection `json:"ombi"`
	Email              langSection `json:"email"`
	Notifications      langSection `json:"notifications"`
	WelcomeEmails      langSection `json:"welcomeEmails"`
	PasswordResets     langSection `json:"passwordResets"`
	InviteEmails       langSection `json:"inviteEmails"`
	PasswordValidation langSection `json:"passwordValidation"`
	HelpMessages       langSection `json:"helpMessages"`
	JSON               string
}

func (ls *setupLangs) getOptions() [][2]string {
	opts := make([][2]string, len(*ls))
	i := 0
	for key, lang := range *ls {
		opts[i] = [2]string{key, lang.Meta.Name}
		i++
	}
	return opts
}

type langSection map[string]string
type tmpl map[string]string

func templateString(text string, vals tmpl) string {
	start, previousEnd := -1, -1
	out := ""
	for i := range text {
		if text[i] == '{' {
			start = i
			continue
		}
		if start != -1 && text[i] == '}' {
			varName := text[start+1 : i]
			val, ok := vals[varName]
			if !ok {
				start = -1
				continue
			}
			out += text[previousEnd+1:start] + val
			previousEnd = i
			start = -1
		}
	}
	if previousEnd != len(text)-1 {
		out += text[previousEnd+1:]
	}
	return out
}

func (el langSection) template(field string, vals tmpl) string {
	text := el.get(field)
	return templateString(text, vals)
}

func (el langSection) format(field string, vals ...string) string {
	text := el.get(field)
	start, previous := -1, -3
	out := ""
	val := 0
	for i := range text {
		if i == len(text)-2 { // Check if there's even enough space for a {n}
			break
		}
		if text[i:i+3] == "{n}" {
			start = i
			out += text[previous+3:start] + vals[val]
			previous = start
			val++
			if val == len(vals) {
				break
			}
		}
	}
	if previous+2 != len(text)-1 {
		out += text[previous+3:]
	}
	return out
}

func (el langSection) get(field string) string {
	t, ok := el[field]
	if !ok {
		return ""
	}
	return t
}
