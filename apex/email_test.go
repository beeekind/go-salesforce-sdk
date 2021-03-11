// +build integration 

package apex_test

import (
	"testing"

	"github.com/beeekind/go-salesforce-sdk/apex"
	"github.com/beeekind/go-salesforce-sdk/client"
	"github.com/beeekind/go-salesforce-sdk/requests"
)

var req = requests.Base.Sender(client.Must())

func TestSendEmail(t *testing.T) {
	response, err := apex.SendEmail(req, &apex.SingleEmailMessage{
		ToAddresses:       []string{"benjamin.stanley.jones@gmail.com"},
		SenderDisplayName: "Benjamin J",
		Subject:           "Programmattic Send 2",
		HTMLBody:          "<h1>Hello, World!</h1>",
	})

	if err != nil {
		println(err.Error())
		t.FailNow()
	}

	t.Log("success", response.Success)
	t.Log("compile", response.CompileProblem)
	t.Log("exception", response.ExceptionMessage)
	t.Log("stack", response.ExceptionStacktrace)
}
