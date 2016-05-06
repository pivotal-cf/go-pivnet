package commands_test

import (
	"os"
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"
)

const (
	apiPrefix = "/api/v2"
	apiToken  = "some-api-token"
)

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commands Suite")
}

var _ = BeforeSuite(func() {
	commands.OutputWriter = os.Stdout
	commands.Pivnet = commands.PivnetCommand{}
	commands.Pivnet.Format = commands.PrintAsJSON
	commands.Pivnet.APIToken = apiToken
})

func fieldFor(command interface{}, name string) reflect.StructField {
	field, success := reflect.TypeOf(command).FieldByName(name)
	Expect(success).To(BeTrue())
	return field
}

func longTag(f reflect.StructField) string {
	return f.Tag.Get("long")
}

func shortTag(f reflect.StructField) string {
	return f.Tag.Get("short")
}

var command = func(f reflect.StructField) string {
	return f.Tag.Get("command")
}

var isRequired = func(f reflect.StructField) bool {
	return f.Tag.Get("required") == "true"
}
