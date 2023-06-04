package request

import (
	"net/url"
)

var AccessTokenParamKey = "access_token"
var VersionParamKey = "v"
var LangParamKey = "lang"
var DeviceIdParamKey = "device_id"
var AnonymousTokenKey = "anonymous_token"

// Параметры запроса к API VK
// Включает в себя как базовые параметры, которые встречаются во всех методах,
// так и дополнительные параметры.
type Params struct {
	params       url.Values // Мэп параметров запроса
	RemoveBlanks bool       // Нужно ли удалять ключи с пустыми значениями при сериализации, по умолчанию - false, всегда задавайте самостоятельно как при декодировании из URL, так и при создании нового объекта
}

// Возвращает объект параметров
func NewParams() *Params {

	p := &Params{
		params: url.Values{},
	}

	p.Version("5.131")
	p.Lang("en")

	return p
}

// Создает объект параметров из url значений или из произвольного map[string][]string
func NewParamsFromUrl(url url.Values) *Params {
	p := NewParams()
	p.params = url
	return p
}

// Перезаписывает ключ существующих параметров, аналогично url.Values{}.Set(key, val)
func (v Params) Set(key, val string) {
	v.params.Set(key, val)
}

// Возвращает значение параметра по ключу, аналогично url.Values{}.Get(key)
func (v *Params) Get(key string) string {
	return v.params.Get(key)
}

// Возвращает информацию о том, есть ли в параметрах значение по указанному ключу, аналогично url.Values{}.Has(key)
func (v *Params) Has(key string) bool {
	return v.params.Has(key)
}

// Удаляет значение параметра, аналогично url.Values{}.Del(key)
func (v Params) Del(key string) {
	v.params.Del(key)
}

// Сериализует параметры в строку, предварительно вызывая метод params.ComposeValues()
func (v *Params) String() string {
	v.ComposeValues()
	return v.params.Encode()
}

// Устанавливает токен доступа
func (v Params) AccessToken(token string) {
	v.Set(AccessTokenParamKey, token)
}

// Устанавливает анонимный токен
func (v Params) AnonymousToken(token string) {
	v.Set(AnonymousTokenKey, token)
}

// Возвращает параметр токена доступа
func (v *Params) GetAccessToken() string {
	return v.Get(AccessTokenParamKey)
}

// Возвращает анонимный токен
func (v *Params) GetAnonymousToken() string {
	return v.Get(AnonymousTokenKey)
}

// Устанавливает версию API
func (v Params) Version(version string) {
	v.Set(VersionParamKey, version)
}

// Возвращает версию API
func (v *Params) GetVersion() string {
	return v.Get(VersionParamKey)
}

// Устанавливает язык
func (v Params) Lang(lang string) {
	v.Set(LangParamKey, lang)
}

// Возвращает язык
func (v *Params) GetLang() string {
	return v.Get(LangParamKey)
}

// Устанавливает идентификатор устройства
func (v Params) DeviceId(deviceId string) {
	v.Set(DeviceIdParamKey, deviceId)
}

// Возвращает идентификатор устройства
func (v *Params) GetDeviceId() string {
	return v.Get(DeviceIdParamKey)
}

// Удаляет пустые значения параметров, если задана настройка RemoveBlanks.
func (v *Params) ComposeValues() {
	if v.RemoveBlanks {
		v.clearBlanks(v.params)
	}
}

// Удалят из параметров ключи с пустыми значениями
func (v *Params) clearBlanks(params url.Values) {
	for key := range params {
		if params.Get(key) == "" {
			delete(params, key)
		}
	}
}
