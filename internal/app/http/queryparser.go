package http

import (
	"net/url"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/protobuf/proto"

	authv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/auth"
)

type CustomQueryParameterParser struct{}

func (p *CustomQueryParameterParser) Parse(target proto.Message, values url.Values, filter *utilities.DoubleArray) error {
	switch req := target.(type) {
	case *authv1.SendEmailRequest:
		return setSendEmailRequest(values, req)
	case *authv1.SendSMSRequest:
		return setSendSMSRequest(values, req)
	case *authv1.ValidateTokenRequest:
		return setValidateTokenRequest(values, req)
	}

	return (&runtime.DefaultQueryParser{}).Parse(target, values, filter)
}

func setSendEmailRequest(values url.Values, r *authv1.SendEmailRequest) error {
	r.Email = values.Get("email")
	if isVerified := values.Get("is_verified"); isVerified != "true" {
		r.IsVerified = false
	} else {
		r.IsVerified = true
	}

	return nil
}

func setSendSMSRequest(values url.Values, r *authv1.SendSMSRequest) error {
	r.Telephone = values.Get("telephone")
	if isVerified := values.Get("is_verified"); isVerified != "true" {
		r.IsVerified = false
	} else {
		r.IsVerified = true
	}

	return nil
}

func setValidateTokenRequest(values url.Values, r *authv1.ValidateTokenRequest) error {
	r.Token = values.Get("token")

	return nil
}
