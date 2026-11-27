package handlers

import (
	"strings"
	"testing"
)

func fptr(f float64) *float64 { return &f }

func TestNormalizeTemplateFields(t *testing.T) {
	got, err := normalizeTemplateFields([]TemplateField{
		{Key: "env", Label: " Environment ", Type: FieldSelect, Options: []string{"dev", " dev", "", "prod"}, Min: fptr(1)},
		{Key: "ram", Label: "RAM", Type: FieldNumber, Options: []string{"x"}, Multiple: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Label != "Environment" || strings.Join(got[0].Options, ",") != "dev,prod" || got[0].Min != nil {
		t.Errorf("select not normalized: %+v", got[0])
	}
	if got[1].Options != nil || got[1].Multiple {
		t.Errorf("number kept select attributes: %+v", got[1])
	}

	bad := [][]TemplateField{
		{{Key: "a", Label: "", Type: FieldText}},
		{{Key: "A b", Label: "A", Type: FieldText}},
		{{Key: "a", Label: "A", Type: FieldText}, {Key: "a", Label: "B", Type: FieldText}},
		{{Key: "a", Label: "A", Type: FieldSelect}},
		{{Key: "a", Label: "A", Type: FieldNumber, Min: fptr(5), Max: fptr(1)}},
		{{Key: "a", Label: "A", Type: "date"}},
	}
	for i, fields := range bad {
		if _, err := normalizeTemplateFields(fields); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

var ec2Fields = []TemplateField{
	{Key: "reason", Label: "Why is it needed", Type: FieldTextarea, Required: true},
	{Key: "ram", Label: "RAM (GB)", Type: FieldNumber, Required: true, Min: fptr(1), Max: fptr(512)},
	{Key: "env", Label: "Account", Type: FieldSelect, Required: true, Options: []string{"dev", "kite"}},
	{Key: "ports", Label: "Ports", Type: FieldSelect, Multiple: true, Options: []string{"22", "443"}},
	{Key: "note", Label: "Note", Type: FieldText},
}

func TestCheckTemplateValues(t *testing.T) {
	_, err := checkTemplateValues(ec2Fields, map[string]any{"ram": "1024", "env": "prod", "extra": "x"})
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{`"Why is it needed" is required`, `"RAM (GB)" must be at most 512`, `"Account" has invalid option "prod"`, `unknown field "extra"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q missing %q", err, want)
		}
	}

	got, err := checkTemplateValues(ec2Fields, map[string]any{
		"reason": "load test", "ram": float64(16), "env": "dev", "ports": []any{"443", "443", "22"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got[1].values[0] != "16" || strings.Join(got[3].values, ",") != "443,22" || len(got[4].values) != 0 {
		t.Errorf("unexpected values: %+v", got)
	}
}

func TestRenderTemplateDescription(t *testing.T) {
	fields := []renderedField{
		{label: "Why <b>", values: []string{"line1\n<script>"}},
		{label: "Ports", values: []string{"22", "443"}, multi: true},
		{label: "Note"},
	}

	h := renderTemplateDescription(fields, "<p>extra</p>")
	for _, want := range []string{"<h3>Why &lt;b&gt;</h3>", "<p>line1<br>&lt;script&gt;</p>", "<li><p>443</p></li>", "<em>No response</em>", "<h3>Additional details</h3><p>extra</p>"} {
		if !strings.Contains(h, want) {
			t.Errorf("html %q missing %q", h, want)
		}
	}
	if strings.Contains(renderTemplateDescription(fields, "<p></p>"), "Additional details") {
		t.Error("empty editor content should not add an extra section")
	}

	md := renderTemplateDescription(fields, "plain **markdown**")
	if !strings.HasPrefix(md, "### Why <b>\n\nline1") || !strings.Contains(md, "- 443\n") || !strings.HasSuffix(md, "### Additional details\n\nplain **markdown**") {
		t.Errorf("unexpected markdown: %q", md)
	}
}
