package lib

import (
	"bytes"
	// "fmt"
	"testing"
)

func TestGetMostWantedItems(t *testing.T) {
	store := NewJSONStore()

	n1 := store.GetItem("n1")
	n2 := store.GetItem("n2")
	n3 := store.GetItem("n3")
	n4 := store.GetItem("n4")
	n5 := store.GetItem("n5")
	n6 := store.GetItem("n6")

	/*
				-------------------------
				n6 <- n2 <- n4
				         <- n5 <- n3
				-------------------------
		     4 dependent nodes

		    -------------------------
		    n2 <- n4
				   <- n5 <- n3
				-------------------------
				 3 dependent nodes

				-------------------------
				n1 <- n3
				      n5 <- n3
				-------------------------
				 2 dependent nodes

				-------------------------
				n5 <- n3
				-------------------------
				 1 dependent node
	*/

	n3.AddDependency(n1)
	n5.AddDependency(n1)
	n2.AddDependency(n6)
	n4.AddDependency(n2)
	n5.AddDependency(n2)
	n3.AddDependency(n5)

	nodes := GetMostWantedItems(store)

	if nodes[0] != n6 {
		t.Errorf("wrong 1 place, expected %s, got %s", n6.Name, nodes[0].Name)
	}

	if nodes[1] != n2 {
		t.Errorf("wrong 1 place, expected %s, got %s", n2.Name, nodes[1].Name)
	}

	if nodes[2] != n1 {
		t.Errorf("wrong 1 place, expected %s, got %s", n1.Name, nodes[2].Name)
	}

	if nodes[3] != n5 {
		t.Errorf("wrong 1 place, expected %s, got %s", n5.Name, nodes[3].Name)
	}

}

func TestGetMostWantedTags(t *testing.T) {
	store := NewJSONStore()

	t1 := store.GetTag("t1")
	t2 := store.GetTag("t2")
	t3 := store.GetTag("t3")
	t4 := store.GetTag("t4")
	t5 := store.GetTag("t5")
	t6 := store.GetTag("t6")

	/*
				-------------------------
				t6 <- t2 <- t4
				         <- t5 <- t3
				-------------------------
		     4 dependent tags

		    -------------------------
		    t2 <- t4
				   <- t5 <- t3
				-------------------------
				 3 dependent tags

				-------------------------
				t1 <- t3
				      t5 <- t3
				-------------------------
				 2 dependent tags

				-------------------------
				t5 <- t3
				-------------------------
				 1 dependent tags
	*/

	t3.AddDependency(t1)
	t5.AddDependency(t1)
	t2.AddDependency(t6)
	t4.AddDependency(t2)
	t5.AddDependency(t2)
	t3.AddDependency(t5)

	tags := GetMostWantedTags(store)

	if tags[0] != t6 {
		t.Errorf("wrong 1 place, expected %s, got %s", t6.Name, tags[0].Name)
	}

	if tags[1] != t2 {
		t.Errorf("wrong 1 place, expected %s, got %s", t2.Name, tags[1].Name)
	}

	if tags[2] != t1 {
		t.Errorf("wrong 1 place, expected %s, got %s", t1.Name, tags[2].Name)
	}

	if tags[3] != t5 {
		t.Errorf("wrong 1 place, expected %s, got %s", t5.Name, tags[3].Name)
	}

}

func TestGetItemsForTags(t *testing.T) {
	store := NewJSONStore()

	n1 := store.GetItem("n1")
	n2 := store.GetItem("n2")
	n3 := store.GetItem("n3")
	t1 := store.GetTag("t1")
	t2 := store.GetTag("t2")
	t3 := store.GetTag("t3")
	n3.AddTag(t3)
	n2.AddTag(t2)
	n1.AddTag(t1)

	if l := len(GetItemsForTags(store, "t1", "t2")); l != 2 {
		t.Errorf("expected %d got %d nodes", 2, l)
	}
}

func TestRemoveTagDependency(t *testing.T) {
	store := NewJSONStore()
	t1 := store.GetTag("t1")
	t2 := store.GetTag("t2")
	t3 := store.GetTag("t3")
	t2.AddDependency(t1)
	t2.AddDependency(t3)

	t2.RemoveDependency(t1.Name)

	if t2.DependsOn[0] != t3.Name {
		t.Errorf("did not remove tag inside dependant tag")
	}

}

func TestRemoveItemDependency(t *testing.T) {
	store := NewJSONStore()

	n1 := store.GetItem("n1")
	n2 := store.GetItem("n2")
	n3 := store.GetItem("n3")
	n2.AddDependency(n1)
	n2.AddDependency(n3)

	n2.RemoveDependency(n1.Name)

	if n2.DependsOn[0] != n3.Name {
		t.Errorf("did not remove node inside dependant node")
	}

}

func TestRemoveTag(t *testing.T) {
	var bf bytes.Buffer
	store := NewJSONStore()
	store.Writer = &bf

	n1 := store.GetItem("n1")
	n2 := store.GetItem("n2")
	t1 := store.GetTag("t1")
	t2 := store.GetTag("t2")
	t2.AddDependency(t1)
	n2.AddTag(t2)
	n1.AddTag(t1)

	store.RemoveTag("t1", true)

	if _, has := store.Tags["t1"]; has {
		t.Errorf("did not remove tag from store")
	}

	if len(n1.Tags) > 0 {
		t.Errorf("did not remove tag inside node")
	}

	if len(t2.DependsOn) > 0 {
		t.Errorf("did not remove tag inside dependant tag")
	}

}

func TestRemoveItem(t *testing.T) {
	var bf bytes.Buffer
	store := NewJSONStore()
	store.Writer = &bf

	n1 := store.GetItem("n1")
	n2 := store.GetItem("n2")
	t1 := store.GetTag("t1")
	t2 := store.GetTag("t2")
	t2.AddDependency(t1)
	n2.AddTag(t2)
	n1.AddTag(t1)

	store.RemoveItem("n1", true)

	if _, has := store.Items["n1"]; has {
		t.Errorf("did not remove node from store")
	}

	if len(n2.DependsOn) > 0 {
		t.Errorf("did not remove node inside dependant node")
	}

}

func TestSaveJSON(t *testing.T) {
	var bf bytes.Buffer
	store := NewJSONStore()
	store.Writer = &bf

	n1 := store.GetItem("n1")
	n2 := store.GetItem("n2")
	n2.AddDependency(n1)
	t1 := store.GetTag("t1")
	t2 := store.GetTag("t2")
	t2.AddDependency(t1)
	n2.AddTag(t2)
	n1.AddTag(t1)

	if err := store.Save(); err != nil {
		t.Errorf("can't save json store: %s", err)
	}

	expected := `{"Items":{"n1":{"Name":"n1","Tags":["t1"]},"n2":{"Name":"n2","Tags":["t2"],"DependsOn":["n1"]}},"Tags":{"t1":{"Name":"t1"},"t2":{"Name":"t2","DependsOn":["t1"]}}}
`
	if bf.String() != expected {
		t.Errorf("saved json string does not match: \n%s\n!=\n%s", bf.String(), expected)
	}
}

func TestLoadJSON(t *testing.T) {
	var bf bytes.Buffer
	bf.WriteString(`
{
	"Items": {
		"n1": {
			"Name": "n1",
			"Tags": ["t1"],
			"DependsOn": []
		},
		"n2": {
			"Name": "n2",
			"Tags": ["t2"],
			"DependsOn": ["n1"]
		}
	},
	"Tags": {
		"t1": {
			"Name": "t1",
			"DependsOn": []
		},
		"t2": {
			"Name": "t2",
			"DependsOn": ["t1"]
		}
	}
}
`)

	store := NewJSONStore()
	store.Reader = &bf
	if err := store.Load(); err != nil {
		t.Errorf("failed to load from JSONStore: %s", err)
	}

	if store.GetItem("n1").Name != "n1" {
		t.Errorf("missing node n1")
	}

	if store.GetItem("n2").Name != "n2" {
		t.Errorf("missing node n2")
	}

	if store.GetItem("n2").DependsOn[0] != "n1" {
		t.Errorf("missing node n2 DependsOn n1")
	}

	if store.GetItem("n2").Tags[0] != "t2" {
		t.Errorf("missing node n2 tag t2")
	}

	if store.GetTag("t1").Name != "t1" {
		t.Errorf("missing tag t1")
	}

	if store.GetTag("t1").Name != "t1" {
		t.Errorf("missing tag t1")
	}

	if store.GetTag("t2").DependsOn[0] != "t1" {
		t.Errorf("missing tag t2 DependsOn t1")
	}
}

func TestRenameItem(t *testing.T) {
	var bf bytes.Buffer

	store := NewJSONStore()
	store.Reader = &bf
	store.Writer = &bf

	n1 := store.GetItem("n1")
	n2 := store.GetItem("n2")
	n3 := store.GetItem("n3")
	t1 := store.GetTag("t1")
	n2.AddTag(t1)
	n2.AddDependency(n3)
	n1.AddDependency(n2)
	RenameItem(store, "n2", "ntwo")

	if n1.DependsOn[0] != store.GetItem("ntwo").Name {
		t.Errorf("renaming node failed: %s != %s", n1.DependsOn[0], store.GetItem("ntwo").Name)
	}

	if store.GetItem("ntwo").Tags[0] != t1.Name {
		t.Errorf("renaming node didn't copy Tags")
	}

	if store.GetItem("ntwo").DependsOn[0] != n3.Name {
		t.Errorf("renaming node didn't copy DependsOn")
	}
}

func TestRenameTag(t *testing.T) {
	var bf bytes.Buffer

	store := NewJSONStore()
	store.Reader = &bf
	store.Writer = &bf

	n1 := store.GetItem("n1")
	t1 := store.GetTag("t1")
	t2 := store.GetTag("t2")
	t3 := store.GetTag("t3")
	t2.AddDependency(t3)
	t1.AddDependency(t2)
	n1.AddTag(t2)

	RenameTag(store, "t2", "ttwo")

	if n1.Tags[0] != store.GetTag("ttwo").Name {
		t.Errorf("renaming tag failed for node tags: %s != %s", n1.Tags[0], store.GetTag("ttwo").Name)
	}

	if t1.DependsOn[0] != store.GetTag("ttwo").Name {
		t.Errorf("renaming tag failed for dependant tags: %s != %s", t1.DependsOn[0], store.GetTag("ntwo").Name)
	}

	if store.GetTag("ttwo").DependsOn[0] != t3.Name {
		t.Errorf("renaming tag didn't copy DependsOn")
	}
}
