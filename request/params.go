package request

import (
	"fmt"
	"net/url"

	"github.com/gorilla/schema"
)

// Параметры запроса к API VK
// Включает в себя как базовые параметры, которые встречаются во всех методах,
// так и дополнительные параметры.
type Params struct {
	Version          string     `schema:"v"`            // Версия API, обязательный параметр
	DeviceId         string     `schema:"device_id"`    // Идентификатор устройства, используется официальными клиентами ВКонтакте
	AccessToken      string     `schema:"access_token"` // Токен доступа аккаунта, используется в большинстве запросов
	Lang             string     `schema:"lang"`         // Язык работы с API
	additionalParams url.Values `schema:"-"`            // Расширенные параметры запроса
	removeBlanks     bool       `schema:"-"`            // Нужно ли удалять ключи с пустыми значениями при сериализации, по умолчанию - false, всегда задавайте самостоятельно как при декодировании из URL, так и при создании нового объекта
}

func NewParams() *Params {
	return &Params{
		Lang:             "ru",
		Version:          "5.141",
		additionalParams: url.Values{},
	}
}

// Создает объект параметров из url значений или из произвольного map[string][]string
func NewParamsFromUrl(url url.Values) (*Params, error) {
	p := NewParams()
	p.additionalParams = url

	err := schema.NewDecoder().Decode(p, url)
	if err != nil {
		return nil, fmt.Errorf("decode params error: %w", err)
	}

	return p, nil
}

// Указывает, нужно ли удалять пустые значения из параметров при сериализации запроса в строку
func (v *Params) RemoveBlanks(remove bool) {
	v.removeBlanks = remove
}

// Возвращает текущее значение RemoveBlanks
func (v *Params) GetRemoveBlanks() bool {
	return v.removeBlanks
}

// Добавляет новый ключ к уже существующим, аналогично url.Values{}.Add(key, val)
func (v *Params) Add(key, val string) {
	v.additionalParams.Add(key, val)
}

// Перезаписывает ключ существующих параметров, аналогично url.Values{}.Set(key, val)
func (v *Params) Set(key, val string) {
	v.additionalParams.Set(key, val)
}

// Сериализует параметры в строку url.Values{}.Encode()
func (v *Params) String() string {
	return v.ComposeValues().Encode()
}

// Собирает глобальные параметры и дополнительные в один map
// Глобальные параметры при этом считаются приоритетными
// Удаляет пустые строки, если RemoveBlanks() == true
func (v *Params) ComposeValues() url.Values {
	params := url.Values{}
	schema.NewEncoder().Encode(v, params)

	for key, val := range v.additionalParams {
		if _, ok := params[key]; !ok {
			params[key] = val
		}
	}

	if v.removeBlanks {
		params = v.clearBlanks(params)
	}

	return params
}

// Удалят из параметров ключи с пустыми значениями
func (v *Params) clearBlanks(params url.Values) url.Values {
	resultParams := url.Values{}

	for key, val := range params {
		if len(val) > 0 {
			for _, v := range val {
				if v != "" {
					resultParams[key] = val
					break
				}
			}
		}
	}

	return resultParams
}
