package pastes

import (
	"reflect"
	"testing"
)

func TestAddChild(t *testing.T) {
	p := New("testdata")
	parent := p.New()
	child := p.New()
	parent.AddChild(child)

	if len(parent.Children) != 1 {
		t.Errorf("%d != 1", len(parent.Children))
	}

	if !parent.Children[child.ID] {
		t.Errorf("Can't find child in parent")
	}

	if child.Parent != parent.ID {
		t.Errorf("%s != %s", child.Parent, parent.ID)
	}

	parent.AddChild(child)

	if len(parent.Children) != 1 {
		t.Errorf("%d != 1", len(parent.Children))
	}

	secondChild := p.New()
	parent.AddChild(secondChild)

	if len(parent.Children) != 2 {
		t.Errorf("%d != 2", len(parent.Children))
	}

	if !parent.Children[secondChild.ID] {
		t.Errorf("Can't find second child in parent")
	}

	if secondChild.Parent != parent.ID {
		t.Errorf("%s != %s", secondChild.Parent, parent.ID)
	}
}

func TestSetParent(t *testing.T) {
	p := New("testdata")
	parent := p.New()
	child := p.New()
	child.SetParent(parent)

	if len(parent.Children) != 1 {
		t.Errorf("%d != 1", len(parent.Children))
	}

	if !parent.Children[child.ID] {
		t.Errorf("Can't find child in parent")
	}

	if child.Parent != parent.ID {
		t.Errorf("%s != %s", child.Parent, parent.ID)
	}

	child.SetParent(parent)

	if len(parent.Children) != 1 {
		t.Errorf("%d != 1", len(parent.Children))
	}

	secondChild := p.New()
	secondChild.SetParent(parent)

	if len(parent.Children) != 2 {
		t.Errorf("%d != 2", len(parent.Children))
	}

	if !parent.Children[secondChild.ID] {
		t.Errorf("Can't find second child in parent")
	}

	if secondChild.Parent != parent.ID {
		t.Errorf("%s != %s", secondChild.Parent, parent.ID)
	}
}

func TestModify(t *testing.T) {
	p := New("testdata")
	parent := p.New()
	parent.Title = "Hello world"
	parent.Content = "Hi there"
	parent.Syntax = "text"

	child := parent.Modify()

	if len(parent.Children) != 1 {
		t.Errorf("%d != 1", len(parent.Children))
	}

	if !parent.Children[child.ID] {
		t.Errorf("Can't find child in parent")
	}

	if child.Parent != parent.ID {
		t.Errorf("%s != %s", child.Parent, parent.ID)
	}

	if !reflect.DeepEqual([]string{child.Title, child.Content, child.Syntax}, []string{parent.Title, parent.Content, parent.Syntax}) {
		t.Errorf("Modify clone doesn't match parent")
	}

	child.Title = "Something else"
	if reflect.DeepEqual([]string{child.Title, child.Content, child.Syntax}, []string{parent.Title, parent.Content, parent.Syntax}) {
		t.Errorf("Modified clone matchs parent")
	}

	secondChild := parent.Modify()

	if len(parent.Children) != 2 {
		t.Errorf("%d != 2", len(parent.Children))
	}

	if !parent.Children[secondChild.ID] {
		t.Errorf("Can't find second child in parent")
	}

	if secondChild.Parent != parent.ID {
		t.Errorf("%s != %s", secondChild.Parent, parent.ID)
	}
}

