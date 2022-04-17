package main

import (
	"net/http"
	"net/http/cookiejar"
	"testing"

	"golang.org/x/net/publicsuffix"
)

func mustGetJar() *cookiejar.Jar {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	if err != nil {
		panic(err)
	}
	return jar

}
func TestClient_GrabSession(t *testing.T) {
	type fields struct {
		Client *http.Client
	}
	type args struct {
		urlStr string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "test",
			fields:  fields{Client: &http.Client{Jar: mustGetJar()}},
			args:    args{urlStr: "https://propaccess.trueautomation.com/clientdb/?cid=56"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Client: tt.fields.Client,
			}
			if err := c.GrabSession(tt.args.urlStr); (err != nil) != tt.wantErr {
				t.Errorf("GrabSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
