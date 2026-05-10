package types

import (
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var urlRegex = regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}(\.[a-z]{2,63})?\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`)

type Url url.URL

func (u *Url) MarshalJSON() ([]byte, error) {
	if u == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(u.URL().String())
}

func (u *Url) UnmarshalJSON(data []byte) error {
	var rawUrl string
	if err := json.Unmarshal(data, &rawUrl); err != nil {
		return err
	}
	URL, err := url.Parse(rawUrl)
	if err != nil {
		return err
	}
	*u = Url(*URL)
	return nil
}

func (u *Url) URL() *url.URL {
	return (*url.URL)(u)
}

func (u *Url) String() string {
	return u.URL().String()
}

func (u *Url) Subpage(subpages ...string) *Url {
	return (*Url)(u.URL().JoinPath(subpages...))
}

func (u *Url) Query() *url.Values {
	vals := u.URL().Query()
	return &vals
}

func (u *Url) SetQuery(query *url.Values) *Url {
	u.URL().RawQuery = query.Encode()
	return u
}

func NewUrl(rawUrl string) (*Url, error) {
	URL, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	return (*Url)(URL), nil
}

func NewUrlSafe(rawUrl string) *Url {
	url, err := NewUrl(rawUrl)
	if err != nil {
		panic(err)
	}
	return url
}

func IsValidUrl(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("url must start with \"http://\" or \"https://\"")
	}
	if !urlRegex.MatchString(url) {
		return errors.New("the url contains illegal characters or is invalid")
	}
	return nil
}

func (url *Url) UnmarshalParam(param string) error {
	if param == "" {
		return errors.New("missing url")
	}
	if err := IsValidUrl(param); err != nil {
		return err
	}
	parsed, err := NewUrl(param)
	if err != nil {
		return errors.New("invalid url")
	}
	*url = *parsed
	return nil
}
