package pastes

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	p := New("testdata")
	if p.path != "testdata" {
		t.Errorf("Path didn't get set in the pastes object")
	}

	if p.rand == nil {
		t.Errorf("Random object wasn't created correctly")
	}
}

func TestNewPaste(t *testing.T) {
	p := New("testdata")
	paste := p.New()

	if paste.ID == "" {
		t.Errorf("Paste.ID didn't get set")
	}

	if paste.pastes == nil {
		t.Errorf("Paste.pastes didn't get set")
	}
}

func TestLoad(t *testing.T) {
	p := New("testdata")
	child, err := p.Load("6LqtsoDl0t0n0vTBiu2Ql6Wj")
	if err != nil {
		t.Errorf("Couldn't load data: %v", err)
	}
	parent, err := child.LoadParent()
	if err != nil {
		t.Errorf("Failed to load child: %v", err)
	}

	childComparison := &Paste{
		pastes: p,
		ID: "6LqtsoDl0t0n0vTBiu2Ql6Wj",
		Content: "<?php\necho \"Hello Better World\";\n?>",
		Syntax: "php",
		Author: "Tester",
		Parent: "7MP4sYDl0t0n7IOq_bK1sfGI",
	}

	parentComparison := &Paste{
		pastes: p,
		ID: "7MP4sYDl0t0n7IOq_bK1sfGI",
		Content: "<?php\necho \"Hello World\";\n?>",
		Syntax: "php",
		Author: "Tester",
		Children: Children{"6LqtsoDl0t0n0vTBiu2Ql6Wj" : true},
	}

	if !reflect.DeepEqual(child, childComparison) {
		t.Errorf("Child != Comparison")
	}

	if !reflect.DeepEqual(parent, parentComparison) {
		t.Errorf("Parent != Comparison")
	}
}

func BenchmarkGenerateID(b *testing.B) {
	p := New("testdata")

	for i := 0; i < b.N; i++ {
		_ = p.GenerateID()
	}
}
