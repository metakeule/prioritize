package webserver

import (
	"encoding/json"
	// "fmt"
	"net/http"

	"lib"
)

type storeServer struct {
	store lib.Store
	name  string
}

// TODO: implement RemoveItem, RemoveTag, RenameItem, RenameTag

type edge struct {
	From string
	To   string
}

func (s *storeServer) RemoveItem(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if req.Method != "DELETE" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var str struct{ Name string }

	if err := json.NewDecoder(req.Body).Decode(&str); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.store.RemoveItem(str.Name, true)

	if err := s.store.Save(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (s *storeServer) AppName(w http.ResponseWriter, req *http.Request) {
	v := struct {
		Name string
	}{
		Name: s.name,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(v)
}

func (s *storeServer) RemoveItemEdge(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != "DELETE" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var str struct{ From, To string }

	if err := json.NewDecoder(req.Body).Decode(&str); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	i1 := s.store.GetItem(str.From)
	i1.RemoveDependency(str.To)

	if err := s.store.Save(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (s *storeServer) PutItemEdge(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != "PUT" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var e edge
	if err := json.NewDecoder(req.Body).Decode(&e); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	i1 := s.store.GetItem(e.From)
	i2 := s.store.GetItem(e.To)

	if i1 == nil || i2 == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// fmt.Printf("add dependency from %#v to %#v\n", i1.Name, i2.Name)

	i1.AddDependency(i2)

	if err := s.store.Save(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *storeServer) PutItem(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != "PUT" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	/*
		b, _ := ioutil.ReadAll(req.Body)
		fmt.Printf("body: %#v\n", string(b))
		return
	*/
	var item lib.Item
	if err := json.NewDecoder(req.Body).Decode(&item); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	n := s.store.GetItem(item.Name)

	// TODO: check if given tags and dependson items do exist,
	// if not => http.StatusBadRequest
	n.Tags = item.Tags
	n.DependsOn = item.DependsOn

	if err := s.store.Save(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *storeServer) PutTag(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if req.Method != "PUT" {
		w.WriteHeader(http.StatusBadRequest)
	}
	var tag lib.Tag
	if err := json.NewDecoder(req.Body).Decode(&tag); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	t := s.store.GetTag(tag.Name)

	// TODO: check if given dependson tags do exist,
	// if not => http.StatusBadRequest
	t.DependsOn = tag.DependsOn

	if err := s.store.Save(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *storeServer) AllItems(w http.ResponseWriter, req *http.Request) {
	var items []*lib.Item

	s.store.EachItem(func(i *lib.Item) {
		items = append(items, i)
	})
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *storeServer) AllTags(w http.ResponseWriter, req *http.Request) {
	var tags []*lib.Tag

	s.store.EachTag(func(t *lib.Tag) {
		tags = append(tags, t)
	})
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tags); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *storeServer) ItemTree(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	t := lib.ItemTree(s.store)
	if err := json.NewEncoder(w).Encode(t); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *storeServer) RenameItem(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var str struct{ Old, New string }

	if err := json.NewDecoder(req.Body).Decode(&str); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lib.RenameItem(s.store, str.Old, str.New)
}

func (s *storeServer) ItemsGraphviz(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(lib.MakeGraphviz(lib.ItemTree(s.store))))
}

func (s *storeServer) TagTree(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	t := lib.TagTree(s.store)
	if err := json.NewEncoder(w).Encode(t); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *storeServer) ItemsVisDataSet(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lib.MakeItemsVisDataSet(s.store, lib.ItemTree(s.store)))
}

func NewStoreServer(name string, store lib.Store) *storeServer {
	return &storeServer{
		store: store,
		name:  name,
	}
}

/*
func sigmaFromItems(s store) *sigma {
	return nil
}

func sigmaFromTags(s store) *sigma {
	return nil
}

type sigma struct {
	Nodes []sigmaNode
	Edges []sigmaEdge
}

type sigmaNode struct {
	ID    string
	Label string
	X     int
	Y     int
	Size  int
}

type sigmaEdge struct {
	ID     string
	Source string
	Target string
}
*/

/*
{
  "nodes": [
    {
      "id": "n0",
      "label": "A node",
      "x": 0,
      "y": 0,
      "size": 3
    },
    {
      "id": "n1",
      "label": "Another node",
      "x": 3,
      "y": 1,
      "size": 2
    },
    {
      "id": "n2",
      "label": "And a last one",
      "x": 1,
      "y": 3,
      "size": 1
    }
  ],
  "edges": [
    {
      "id": "e0",
      "source": "n0",
      "target": "n1"
    },
    {
      "id": "e1",
      "source": "n1",
      "target": "n2"
    },
    {
      "id": "e2",
      "source": "n2",
      "target": "n0"
    }
  ]
}
*/
