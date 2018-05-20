package config

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/go-test/deep"
	"github.com/spf13/viper"
)

func TestTranslateViperPath(t *testing.T) {
	type testCase struct {
		Prefix       string
		ViperKey     string
		ExpectedPath string
	}

	cases := []testCase{
		testCase{"/test", "foo.bar", "/test/foo/bar"},
		testCase{"test", "foo.bar", "/test/foo/bar"},
		testCase{"/test/", "foo.bar", "/test/foo/bar"},
	}
	for _, c := range cases {
		p := &ParameterViper{parameterPrefix: c.Prefix}
		actualPath := p.translateViperPath(c.ViperKey)
		if actualPath != c.ExpectedPath {
			t.Errorf("Expecting '%s', got '%s': Supplied Prefix '%s' - key '%s'", c.ExpectedPath, actualPath, c.Prefix, c.ViperKey)
		}
	}
}

type testReturnValue struct {
	Parameter *ssm.Parameter
	Err       error
}
type testParameterStore struct {
	Values map[string]testReturnValue
}

func (s *testParameterStore) GetParameter(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	result, ok := s.Values[*in.Name]
	if !ok {
		return nil, awserr.New(ssm.ErrCodeParameterNotFound, "", nil)
	}
	return &ssm.GetParameterOutput{Parameter: result.Parameter}, result.Err
}

func TestGetParameter(t *testing.T) {
	s := &testParameterStore{Values: map[string]testReturnValue{
		"/prefix/testString": testReturnValue{
			Parameter: &ssm.Parameter{
				Type:  aws.String("String"),
				Value: aws.String("TestValue"),
			},
			Err: nil,
		},
	},
	}

	p := ParameterViper{parameterPrefix: "prefix", parameterStore: s}
	param, err := p.getParamater("/prefix/testString")
	if err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if *param.Value != "TestValue" {
		t.Errorf("Returned '%s', expecting 'TestValue'", *param.Value)
	}
}

func TestGetParameterNotFound(t *testing.T) {
	s := &testParameterStore{Values: map[string]testReturnValue{}}

	p := ParameterViper{parameterPrefix: "prefix", parameterStore: s}
	_, err := p.getParamater("/prefix/testString")
	if err != errParameterNotFound {
		t.Errorf("expected errParameterNotFound, got: %v", err)
	}
}

func TestGetStringFromParameterStore(t *testing.T) {
	s := &testParameterStore{Values: map[string]testReturnValue{
		"/prefix/test/foo": testReturnValue{
			Parameter: &ssm.Parameter{
				Type:  aws.String("String"),
				Value: aws.String("TestValue"),
			},
			Err: nil,
		},
	},
	}

	p := ParameterViper{parameterPrefix: "prefix", parameterStore: s, Viper: viper.New()}
	value := p.GetString("test.foo")
	if value != "TestValue" {
		t.Errorf("Returned '%s', expecting 'TestValue'", value)
	}
}

func TestGetStringFromViper(t *testing.T) {
	s := &testParameterStore{Values: map[string]testReturnValue{}}

	p := ParameterViper{parameterPrefix: "prefix", parameterStore: s, Viper: viper.New()}
	p.SetDefault("test.foo", "TestValue")
	value := p.GetString("test.foo")
	if value != "TestValue" {
		t.Errorf("Returned '%s', expecting 'TestValue'", value)
	}
}

func TestGetStringSliceFromParameterStore(t *testing.T) {
	s := &testParameterStore{Values: map[string]testReturnValue{
		"/prefix/test/foo": testReturnValue{
			Parameter: &ssm.Parameter{
				Type:  aws.String("String"),
				Value: aws.String("Bar,Baz,Alice,Bob"),
			},
			Err: nil,
		},
	},
	}

	p := ParameterViper{parameterPrefix: "prefix", parameterStore: s, Viper: viper.New()}
	value := p.GetStringSlice("test.foo")
	expected := []string{"Bar", "Baz", "Alice", "Bob"}
	if diff := deep.Equal(value, expected); len(diff) > 0 {
		t.Errorf("Returned value doesn't match expected:")
		for _, l := range diff {
			t.Error(l)
		}
	}
}

func TestGetStringSliceFromViper(t *testing.T) {
	s := &testParameterStore{Values: map[string]testReturnValue{}}

	p := ParameterViper{parameterPrefix: "prefix", parameterStore: s, Viper: viper.New()}
	expected := []string{"Bar", "Baz", "Alice", "Bob"}
	p.SetDefault("test.foo", expected)
	value := p.GetStringSlice("test.foo")
	if diff := deep.Equal(value, expected); len(diff) > 0 {
		t.Errorf("Returned value doesn't match expected:")
		for _, l := range diff {
			t.Error(l)
		}
	}
}
