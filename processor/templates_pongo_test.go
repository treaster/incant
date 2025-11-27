package processor_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treaster/incant/processor"
)

func TestBasicPongo(t *testing.T) {
	templateMgr := processor.PongoTemplateMgr("data_url_root/")

	tmpl1 := []byte(`
	1. Constant
	2. {{ data.Value }}
	3. {{ data.Nested.Value }}
	4. {{ Sprintf("ab %q ef", "cd") }}
	<div attr="value" />
	`)

	err := templateMgr.ParseOne("template_1", tmpl1)
	require.NoError(t, err)

	input := struct {
		Value  string
		Nested struct {
			Value int
		}
	}{
		"abc",
		struct {
			Value int
		}{
			5,
		},
	}

	var output bytes.Buffer
	err = templateMgr.Execute("template_1", input, &output)
	require.NoError(t, err)

	require.Equal(t, `
	1. Constant
	2. abc
	3. 5
	4. ab "cd" ef
	<div attr="value" />
	`,
		output.String())
}

func TestMultifilePongo(t *testing.T) {
	templateMgr := processor.PongoTemplateMgr("data_url_root/")

	tmpl1 := []byte(`
	HEADER
	{% block content %}{% endblock %}
	FOOTER
	`)

	tmpl2 := []byte(`
	{% extends "template_1.pongo" %}
	{% block content %}
	CONTENT
	{% endblock %}
	`)

	err := templateMgr.ParseOne("template_1.pongo", tmpl1)
	require.NoError(t, err)

	err = templateMgr.ParseOne("template_2.pongo", tmpl2)
	require.NoError(t, err)

	input := struct{}{}

	var output bytes.Buffer
	err = templateMgr.Execute("template_2.pongo", input, &output)
	require.NoError(t, err)

	require.Equal(t, `
	HEADER
	
	CONTENT
	
	FOOTER
	`,
		output.String())

}
