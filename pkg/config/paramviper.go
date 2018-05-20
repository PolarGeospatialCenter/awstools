package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/viper"
)

var errParameterNotFound = errors.New("Value not found in parameter store")

type parameterGetter interface {
	GetParameter(*ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
}

// ParameterViper wraps a Viper instance.  First attempting to find the corresponding entry in ssm parameter store
type ParameterViper struct {
	parameterPrefix string
	parameterStore  parameterGetter
	*viper.Viper
}

// NewParameterViper creates a new ParameterViper using the value of the 'PARAMETER_STORE_PREFIX' environment variable as the prefix
func NewParameterViper() *ParameterViper {
	prefix := os.Getenv("PARAMETER_STORE_PREFIX")
	if prefix == "" {
		prefix = "parameters"
	}

	p := &ParameterViper{parameterPrefix: prefix, Viper: viper.New()}
	p.parameterStore = ssm.New(session.New())
	return p
}

// SetParameterStorePrefix sets the prefix to use with the parameter store
func (p *ParameterViper) SetParameterStorePrefix(prefix string) {
	p.parameterPrefix = prefix
}

// translatePath replaces '.' in the viper path with '/' and prepend the ParameterPrefix
func (p *ParameterViper) translateViperPath(viperPath string) string {
	viperParts := strings.Split(viperPath, ".")
	prefixParts := strings.Split(p.parameterPrefix, "/")
	parameterParts := append(prefixParts, viperParts...)
	return "/" + path.Join(parameterParts...)
}

func (p *ParameterViper) getParamater(path string) (*ssm.Parameter, error) {
	out, err := p.parameterStore.GetParameter(&ssm.GetParameterInput{Name: aws.String(path), WithDecryption: aws.Bool(true)})
	if err != nil {
		log.Printf("unable to read parameter from parameter store: %v", err)
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeInternalServerError:
				return nil, fmt.Errorf("parameter store returned internal server error: %v", err)
			default:
				return nil, errParameterNotFound
			}
		}
		return nil, fmt.Errorf("unknown error retrieving parameter: %v", err)
	}
	return out.Parameter, nil
}

// GetString first looks for a string in the aws parameter store, then falls back to the viper store
func (p *ParameterViper) GetString(key string) string {
	param, err := p.getParamater(p.translateViperPath(key))
	if err != nil {
		return p.Viper.GetString(key)
	}
	return *param.Value
}

// GetStringSlice first looks for a string map in the aws parameter store, then falls back to the viper store
func (p *ParameterViper) GetStringSlice(key string) []string {
	param, err := p.getParamater(p.translateViperPath(key))
	if err != nil {
		return p.Viper.GetStringSlice(key)
	}
	return strings.Split(*param.Value, ",")
}
