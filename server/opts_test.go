package server_test

import (
	"github.com/RobinUS2/tsxdb/server"
	"github.com/RobinUS2/tsxdb/tools"
	"io/ioutil"
	"testing"
)

func TestOpts_ReadYamlFile(t *testing.T) {
	opts := server.NewOpts()
	// clear
	opts.ListenHost = ""
	opts.ListenPort = 0
	opts.AuthToken = ""
	tmpFile, err := ioutil.TempFile("", "opts_test")
	if err != nil {
		t.Error(err)
	}
	yml := `
connection:
  listen_port: 1234
  listen_host: 0.0.0.0
  auth_token: "verySecure"
`
	if err := ioutil.WriteFile(tmpFile.Name(), []byte(yml), 0644); err != nil {
		t.Error(err)
	}
	if err := tools.ReadYamlFile(tmpFile.Name(), &opts); err != nil {
		t.Error(err)
	}
	if opts.ListenPort != 1234 {
		t.Error(opts.ListenPort)
	}
	if opts.ListenHost != "0.0.0.0" {
		t.Error(opts.ListenHost)
	}
	if opts.AuthToken != "verySecure" {
		t.Error(opts.AuthToken)
	}
}
