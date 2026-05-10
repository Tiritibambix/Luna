package util

import (
	"encoding/json"
	"fmt"
	"luna-backend/auth"
	"luna-backend/constants"
	"luna-backend/errors"
	"luna-backend/protocols"
	"luna-backend/protocols/caldav"
	"luna-backend/protocols/google"
	"luna-backend/protocols/ical"
	"luna-backend/types"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	ContentJson = iota
	ContentForm
	ContentFormMultipart
)

func determineContentType(c *gin.Context) int {
	switch c.ContentType() {
	case "application/x-www-form-urlencoded":
		return ContentForm
	case "multipart/form-data":
		return ContentFormMultipart
	case "application/json":
		fallthrough
	default:
		return ContentJson
	}
}

func ParseIntoMap(c *gin.Context) (map[string]json.RawMessage, *errors.ErrorTrace) {
	var values url.Values
	mapped := map[string]json.RawMessage{}

	switch determineContentType(c) {
	case ContentJson:
		return nil, ParseIntoObject(c, mapped)
	case ContentForm:
		err := c.Request.ParseForm()
		if err != nil {
			return nil, errors.New().Status(http.StatusBadRequest).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlWordy, "Could not parse form data").
				AltStr(errors.LvlPlain, "Malformed form data")
		}
		values = c.Request.Form
	case ContentFormMultipart:
		err := c.Request.ParseMultipartForm(constants.MaxFormBytes)
		if err != nil {
			return nil, errors.New().Status(http.StatusBadRequest).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlWordy, "Could not parse multipart form data").
				AltStr(errors.LvlPlain, "Malformed form data")
		}
		values = c.Request.PostForm
	}

	for key, value := range values {
		if len(value) == 1 {
			mapped[key] = []byte(value[0])
		} else {
			mapped[key] = fmt.Appendf(nil, "[%s]", strings.Join(value, ","))
		}
	}

	return mapped, nil
}

func ParseIntoObject(c *gin.Context, obj any) *errors.ErrorTrace {
	var embeddedJson json.RawMessage

	var err error
	switch determineContentType(c) {
	case ContentJson:
		err = c.ShouldBindBodyWithJSON(obj)
	case ContentForm:
		err = c.ShouldBindWith(obj, binding.Form)
		embeddedJson = []byte(c.Request.Form.Get("json"))
	case ContentFormMultipart:
		err = c.ShouldBindWith(obj, binding.FormMultipart)
		embeddedJson = []byte(c.Request.PostForm.Get("json"))
	}

	// Ignore validation errors if we are not done unmarshaling yet
	if len(embeddedJson) != 0 && (err == nil || strings.Contains(err.Error(), "Field validation")) {
		err = json.Unmarshal(embeddedJson, obj)
	}

	if err != nil {
		return errors.New().Status(http.StatusBadRequest).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not unmarshal body").
			Append(errors.LvlPlain, "Malformed request")
	}

	return nil
}

func ParseAuth(c *gin.Context, authType string, userId types.ID) (types.AuthMethod, *errors.ErrorTrace) {
	authMethod, err := auth.EmptyAuthByType(authType)
	if err != nil {
		return nil, errors.New().
			Status(http.StatusBadRequest).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not parse auth method").
			Append(errors.LvlPlain, "Malformed request")
	}

	if oauthAuth, ok := authMethod.(*auth.OauthAuth); ok {
		oauthAuth.UserId = userId
	}

	var tr *errors.ErrorTrace
	switch authMethod := authMethod.(type) {
	case *auth.NoAuth:
		tr = ParseIntoObject(c, &struct {
			Auth *auth.NoAuth `json:"auth"`
		}{authMethod})
	case *auth.BasicAuth:
		tr = ParseIntoObject(c, &struct {
			Auth *auth.BasicAuth `json:"auth"`
		}{authMethod})
	case *auth.BearerAuth:
		tr = ParseIntoObject(c, &struct {
			Auth *auth.BearerAuth `json:"auth"`
		}{authMethod})
	case *auth.OauthAuth:
		tr = ParseIntoObject(c, &struct {
			Auth *auth.OauthAuth `json:"auth"`
		}{authMethod})
	}
	if tr != nil {
		return nil, tr
	}

	return authMethod, nil
}

func ParseSourceSettings(c *gin.Context, sourceType string) (types.SourceSettings, *errors.ErrorTrace) {
	sourceSettings, err := protocols.EmptySourceSettingsByType(sourceType)
	if err != nil {
		return nil, errors.New().
			Status(http.StatusBadRequest).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not parse source settings").
			Append(errors.LvlPlain, "Malformed request")
	}

	var obj any
	switch sourceSettings := sourceSettings.(type) {
	case *caldav.CaldavSourceSettings:
		obj = struct {
			Settings *caldav.CaldavSourceSettings `json:"source"`
		}{sourceSettings}
	case *google.GoogleSourceSettings:
		obj = struct {
			Settings *google.GoogleSourceSettings `json:"source"`
		}{sourceSettings}
	case *ical.IcalSourceSettings:
		obj = struct {
			Settings *ical.IcalSourceSettings `json:"source"`
		}{sourceSettings}
	}

	tr := ParseIntoObject(c, obj)
	if tr != nil {
		return nil, tr
	}

	return sourceSettings, nil
}
