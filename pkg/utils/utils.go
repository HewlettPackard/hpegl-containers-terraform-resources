package utils

import "github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"

func GetErrorMessage(err error, statusCode int) string {
	swaggerErr, ok := err.(mcaasapi.GenericSwaggerError)
	if !ok {
		return ""
	}

	model := swaggerErr.Model()
	switch statusCode {
	case 400:
		badRequestModel, ok := model.(mcaasapi.ModelError)
		if !ok {
			return ""
		}
		return badRequestModel.Message
	case 401:
		authErrorModel, ok := model.(mcaasapi.ModelError)
		if !ok {
			return ""
		}
		return authErrorModel.Message
	case 422:
		unprocessingEntityModel, ok := model.(mcaasapi.ModelError)
		if !ok {
			return ""
		}
		return unprocessingEntityModel.Message
	case 500:
		internalErrorModel, ok := model.(mcaasapi.ModelError)
		if !ok {
			return ""
		}
		return internalErrorModel.Message
	default:
		return ""
	}
}
