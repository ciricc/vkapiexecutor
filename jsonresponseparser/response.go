package jsonresponseparser

import (
	"encoding/json"
	"net/http"

	"github.com/buger/jsonparser"
	response "github.com/ciricc/vkapiexecutor/response"
)

// Объект ответа VK API в формате JSON
type JsonResponse struct {
	response.UnknownResponse
}

func NewApiJsonResponse(httpResponse *http.Response) *JsonResponse {
	res := &JsonResponse{
		*response.NewUnknown(httpResponse),
	}
	return res
}

// Валидирует JSON
func (v *JsonResponse) ValidateJson() error {
	t := map[string]interface{}{}
	return json.Unmarshal(v.Body(), &t)
}

// Возвращает информацию об ошибке выполнения метода
func (v *JsonResponse) Error() error {
	body := v.Body()

	errorObject, _, _, err := jsonparser.Get(body, "error")
	if err != nil {
		return nil
	}

	errorMessage, _ := jsonparser.GetString(errorObject, "error_msg")
	errorMessageIntCode, _ := jsonparser.GetInt(errorObject, "error_code")

	if errorMessage == "" && errorMessageIntCode == 0 {
		return nil
	}

	apiError := response.NewError(errorMessage, int(errorMessageIntCode))

	redirectUri, err := jsonparser.GetString(errorObject, "redirect_uri")
	if err == nil {
		apiError.RedirectUri = redirectUri
	}

	captchaImg, err := jsonparser.GetString(errorObject, "captcha_img")
	if err == nil {
		apiError.CaptchaImg = captchaImg
	}

	captchaSid, err := jsonparser.GetString(errorObject, "captcha_sid")
	if err == nil {
		apiError.CaptchaSid = captchaSid
	}

	return apiError
}
