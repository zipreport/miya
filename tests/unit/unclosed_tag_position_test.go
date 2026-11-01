package miya_test

import (
	"strings"
	"testing"

	"github.com/zipreport/miya"
)

func TestUnclosedTagPosition(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"if", "x\n{% if a %}y", "at line 2, column "},
		{"for", "x\n{% for a in b %}y", "at line 2, column "},
		{"block", "x\n{% block b %}y", "at line 2, column "},
		{"with", "x\n{% with a = 1 %}y", "at line 2, column "},
		{"macro", "x\n{% macro m() %}y", "at line 2, column "},
		{"filter", "x\n{% filter upper %}y", "at line 2, column "},
		{"autoescape", "x\n{% autoescape true %}y", "at line 2, column "},
		{"set_block", "x\n{% set a %}y", "at line 2, column "},
		{"call", "x\n{% call m() %}y", "at line 2, column "},
		{"three_lines", "a\nb\n{% if x %}\nc", "at line 4, column "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := miya.NewEnvironment().FromString(tc.src)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not contain %q", err, tc.want)
			}
			if strings.Contains(err.Error(), "line 0") {
				t.Errorf("error %q mentions line 0", err)
			}
		})
	}
}
