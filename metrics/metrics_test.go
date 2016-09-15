package metrics

import "testing"

func TestEventString(t *testing.T) {
	e := &Event{
		Name: "foo",
		Labels: map[string]string{
			"volume":   "baz",
			"instance": "qux",
		},
		Value: "bar",
	}
	expected := "foo{volume=\"baz\",instance=\"qux\"} bar"
	if e.String() != expected {
		t.Fatalf("Expected %s, got %s", expected, e.String())
	}
}

func TestNewMetrics(t *testing.T) {
	p := NewMetrics("foo", "http://foo:9091")

	if p.Instance != "foo" {
		t.Fatalf("Expected instance to be foo, got %s", p.Instance)
	}

	if p.PushgatewayURL != "http://foo:9091" {
		t.Fatalf("Expected URL to be http://foo:9091, got %s", p.PushgatewayURL)
	}

	if len(p.Metrics) != 0 {
		t.Fatalf("Expected empty Metrics array, got size %v", len(p.Metrics))
	}
}

func TestNewMetric(t *testing.T) {
	p := NewMetrics("foo", "http://foo:9091")
	m := p.NewMetric("bar", "qux")

	if len(p.Metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %v", len(p.Metrics))
	}

	if p.Metrics["bar"] != m {
		t.Fatal("Expected to find metric in handler")
	}

	if m.Name != "bar" {
		t.Fatalf("Expected name to be bar, got %s", m.Name)
	}

	if m.Type != "qux" {
		t.Fatalf("Expected type to be qux, got %s", m.Name)
	}
}
