package request_test

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/ciricc/vkapiexecutor/request"
)

func changeAndValidateParams(t *testing.T, expectedVal string, changer func(v string), validator func() string) {
	changer(expectedVal)
	if validator() != expectedVal {
		t.Errorf("parameter was changed by %v changed into %s value, but nothing changed in %v validator", reflect.TypeOf(changer), expectedVal, reflect.TypeOf(validator))
	}
}
func TestParams(t *testing.T) {
	params := request.NewParams()
	params.RemoveBlanks = false

	t.Run("params encoding", func(t *testing.T) {
		queryStringParams := params.String()
		values, err := url.ParseQuery(queryStringParams)
		if err != nil {
			t.Errorf("parse encoded query error: %s", err)
		}

		if values.Encode() != queryStringParams {
			t.Errorf("encoded original prameters are not same:\norigin: %q\nencoded: %q\n", queryStringParams, values.Encode())
		}
	})

	t.Run("params changes self", func(t *testing.T) {
		cases := [](func(*request.Params) error){
			func(p *request.Params) error {
				changeAndValidateParams(t, "access_tokein", p.AccessToken, p.GetAccessToken)
				return fmt.Errorf("access token same")
			},
			func(p *request.Params) error {
				changeAndValidateParams(t, "5.130", p.Version, p.GetVersion)
				return fmt.Errorf("version same")
			},
			func(p *request.Params) error {
				changeAndValidateParams(t, "fr", p.Lang, p.GetLang)
				return fmt.Errorf("langage same")
			},
			func(p *request.Params) error {
				changeAndValidateParams(t, "1234567890", p.DeviceId, p.GetDeviceId)
				return fmt.Errorf("device id same")
			},
			func(p *request.Params) error {
				changeAndValidateParams(t, "anonymous_token", p.AnonymousToken, p.GetAnonymousToken)
				return fmt.Errorf("anonymous token same")
			},
		}

		for _, testCase := range cases {
			oldParams := params.String()
			err := testCase(params)
			if params.String() == oldParams {
				t.Errorf("check chaining difference got same parameters:\norigin: %q\nchanged: %q\ncase: %s", oldParams, params.String(), err)
			}
		}
	})

	t.Run("params decoding from custom url values", func(t *testing.T) {
		expectedVersion := "5.130"
		expectedAccessToken := "abc"
		expectedLanguage := "fr"
		expectedDeviceId := "1234567"
		encodedValues := url.Values{
			request.VersionParamKey:     {expectedVersion},
			request.DeviceIdParamKey:    {expectedDeviceId},
			request.AccessTokenParamKey: {expectedAccessToken},
			request.LangParamKey:        {expectedLanguage},
		}
		params := request.NewParamsFromUrl(encodedValues)

		if params.GetAccessToken() != expectedAccessToken {
			t.Errorf("decoded access token: %q\nbut epected: %q", params.GetAccessToken(), expectedAccessToken)
		}

		if params.GetDeviceId() != expectedDeviceId {
			t.Errorf("decoded access device id: %q\nbut epected: %q", params.GetDeviceId(), expectedDeviceId)
		}

		if params.GetVersion() != expectedVersion {
			t.Errorf("decoded version: %q\nbut epected: %q", params.GetVersion(), expectedVersion)
		}

		if params.GetLang() != expectedLanguage {
			t.Errorf("decoded language: %q\nbut epected: %q", params.GetLang(), expectedLanguage)
		}
	})

	t.Run("default global params tokens", func(t *testing.T) {
		cases := [][]string{
			{request.AccessTokenParamKey, "access_token"},
			{request.VersionParamKey, "v"},
			{request.LangParamKey, "lang"},
			{request.DeviceIdParamKey, "device_id"},
		}
		for _, testCase := range cases {
			if testCase[0] != testCase[1] {
				t.Errorf("expected defalt global param: %q, but it is %q", testCase[1], testCase[0])
			}
		}
	})

	t.Run("same compose values and string value", func(t *testing.T) {
		original := params.String()
		params.ComposeValues()
		composed := params.String()
		if original != composed {
			t.Errorf("not same composed values encoded and string fromatted params")
		}
	})

	t.Run("compose clearing blanks in original params", func(t *testing.T) {
		params := request.NewParams()
		params.RemoveBlanks = true

		params.Set("key", "")
		if !params.Has("key") {
			t.Errorf("key does not exist")
		}

		params.ComposeValues()
		if params.Has("key") {
			t.Errorf("key exist after clear blanks")
		}
	})

	t.Run("delete param", func(t *testing.T) {
		params := request.NewParams()
		params.AccessToken("1")
		params.Del(request.AccessTokenParamKey)
		if params.Has(request.AccessTokenParamKey) {
			t.Errorf("delete not working")
		}
	})

	t.Run("params from url wotking correctly", func(t *testing.T) {
		rawUrl := "https://api.vk.com/method/users.getById?user_ids=1&access_token=abc&v=5.81"
		url, err := url.Parse(rawUrl)
		if err != nil {
			t.Error(err)
		}

		params := request.NewParamsFromUrl(url.Query())
		if params.GetAccessToken() != "abc" {
			t.Errorf("got access token is not abc: %q", params.GetAccessToken())
		}

		if params.GetVersion() != "5.81" {
			t.Errorf("got version is not 5.81: %q", params.GetVersion())
		}

		if !params.Has("user_ids") {
			t.Errorf("not found custom parameter")
		}
	})

	t.Run("params using default global access token key", func(t *testing.T) {
		globalKey := request.AccessTokenParamKey
		defer func() {
			request.AccessTokenParamKey = globalKey
		}()

		request.AccessTokenParamKey = "access_key"

		params := request.NewParams()
		params.Set("access_key", "123")
		if params.GetAccessToken() != "123" {
			t.Errorf("not changed key for access token")
		}
	})

	t.Run("params using default global version key", func(t *testing.T) {
		globalKey := request.VersionParamKey
		defer func() {
			request.VersionParamKey = globalKey
		}()

		request.VersionParamKey = "version"
		params := request.NewParams()
		params.Set("version", "5.74")

		if params.GetVersion() != "5.74" {
			t.Errorf("not changed key for version")
		}
	})

	t.Run("params using default global language key", func(t *testing.T) {
		globalKey := request.LangParamKey
		defer func() {
			request.LangParamKey = globalKey
		}()

		request.LangParamKey = "language"
		params := request.NewParams()
		params.Set("language", "ru")

		if params.GetLang() != "ru" {
			t.Errorf("not changed key for lang")
		}
	})

	t.Run("params using default global device id key", func(t *testing.T) {
		globalKey := request.DeviceIdParamKey
		defer func() {
			request.DeviceIdParamKey = globalKey
		}()

		request.DeviceIdParamKey = "device_identifier"
		params := request.NewParams()
		params.Set("device_identifier", "123456")

		if params.GetDeviceId() != "123456" {
			t.Errorf("not changed key for device id")
		}
	})

	t.Run("params using default global device id key after changing key", func(t *testing.T) {
		globalKey := request.DeviceIdParamKey
		defer func() {
			request.DeviceIdParamKey = globalKey
		}()

		params := request.NewParams()
		request.DeviceIdParamKey = "device_identifier"
		params.Set("device_identifier", "123456")

		if params.GetDeviceId() != "123456" {
			t.Errorf("not changed key for device id")
		}
	})
}
