// THIS FILE IS AUTOGENERATED; DO NOT EDIT

package main

import "github.com/wercker/tracery/tracery"

// TracingTraceme is an autogenerated wrapper for the Traceme type
type TracingTraceme struct {
	tracery.Traceme
	Hero string
}

//{{define "method"}}
func (_t *TracingTraceme) GetObj(methodInput tracery.GetObjInput) tracery.GetObjOutput {
	methodOutput := _t.Traceme.GetObj(methodInput)
	return methodOutput
}

//{{end}}
//{{range .Methods}}{{template "method" .}}{{end}}

var _ tracery.Traceme = (*TracingTraceme)(nil)
