package lib

import (
	"encoding/json"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Store interface {
	Load() (err error)

	GetItem(name string) *Item
	RemoveItem(name string, removeReferences bool)
	EachItem(func(*Item))

	GetTag(name string) *Tag
	RemoveTag(name string, removeReferences bool)
	EachTag(func(*Tag))

	Save() error
}

func NewJSONStore() *JSONStore {
	return &JSONStore{
		Items: map[string]*Item{},
		Tags:  map[string]*Tag{},
	}
}

type JSONStore struct {
	mx     sync.Mutex `json:"-"`
	Items  map[string]*Item
	Tags   map[string]*Tag
	Reader io.Reader `json:"-"`
	Writer io.Writer `json:"-"`
}

func (j *JSONStore) Load() error {
	j.mx.Lock()
	defer j.mx.Unlock()
	if s, is := j.Reader.(io.Seeker); is {
		s.Seek(0, 0)
	}
	return json.NewDecoder(j.Reader).Decode(j)
}

func (j *JSONStore) Save() error {
	j.mx.Lock()
	defer j.mx.Unlock()
	if s, is := j.Writer.(io.Seeker); is {
		_, err := s.Seek(0, 0)
		if err != nil {
			return err
		}
		b, err := json.MarshalIndent(j, "", "    ")
		if err != nil {
			return err
		}

		if size, err := j.Writer.Write(b); err != nil && err != io.EOF {
			// fmt.Printf("write error: %s\n", err.Error())
			return err
		} else {
			if t, is := j.Writer.(interface {
				Truncate(int64) error
			}); is {
				t.Truncate(int64(size))
			}
		}

		if sy, is := j.Writer.(interface {
			Sync() error
		}); is {
			sy.Sync()
		}
		return nil
	}
	return json.NewEncoder(j.Writer).Encode(j)
}

func (j *JSONStore) GetItem(name string) *Item {
	j.mx.Lock()
	defer j.mx.Unlock()
	if n, has := j.Items[name]; has {
		return n
	}

	n := &Item{Name: name}

	j.Items[name] = n
	return n
}

func (j *JSONStore) EachItem(fn func(*Item)) {
	for _, n := range j.Items {
		fn(n)
	}
}

func (j *JSONStore) EachTag(fn func(*Tag)) {
	for _, t := range j.Tags {
		fn(t)
	}
}

// If removeReferences is true: remove all references inside other items
func (j *JSONStore) RemoveItem(name string, removeReferences bool) {
	j.mx.Lock()
	delete(j.Items, name)
	j.mx.Unlock()
	if removeReferences {
		j.EachItem(func(n *Item) {
			n.RemoveDependency(name)
		})
	}
}

func (j *JSONStore) GetTag(name string) *Tag {
	j.mx.Lock()
	defer j.mx.Unlock()
	if t, has := j.Tags[name]; has {
		return t
	}

	t := &Tag{Name: name}

	j.Tags[name] = t
	return t
}

// If removeReferences is true: remove all references inside other tags and items
func (j *JSONStore) RemoveTag(name string, removeReferences bool) {
	j.mx.Lock()
	delete(j.Tags, name)
	j.mx.Unlock()
	if removeReferences {
		j.EachTag(func(t *Tag) {
			t.RemoveDependency(name)
		})
		j.EachItem(func(n *Item) {
			n.RemoveTag(name)
		})
	}
}

type Tag struct {
	Name      string
	DependsOn []string `json:",omitempty"`
}

func (t *Tag) isDependingOn(store Store, other *Tag, visited map[*Tag]bool) (hops int32) {
	if visited[other] {
		return -1
	}

	visited[t] = true

	if t == other {
		return 0
	}

	for _, d := range t.DependsOn {
		dt := store.GetTag(d)
		if !visited[dt] {

			h := dt.isDependingOn(store, other, visited)
			if h == 0 {
				return 1
			}

			if h > 0 {
				return h + 1
			}
		}
	}

	return -1

}

// IsDependingOn checks if t depends on other and returns the number of hops
// 0 is returned if t == other; -1 is returned if t does not depend other
func (t *Tag) IsDependingOn(store Store, other *Tag) (hops int32) {
	return t.isDependingOn(store, other, map[*Tag]bool{})
}

// AddDependency does nothing if the dependency tag has the same name as the current tag
func (t *Tag) AddDependency(d *Tag) {
	if t.Name != d.Name {
		t.DependsOn = append(t.DependsOn, d.Name)
	}
}

func (t *Tag) RemoveDependency(tagName string) {
	var a []string

	for _, d := range t.DependsOn {
		if d != tagName {
			a = append(a, d)
		}
	}

	t.DependsOn = a
}

type Item struct {
	Name      string
	Tags      []string `json:",omitempty"`
	DependsOn []string `json:",omitempty"`
}

func (n *Item) isDependingOn(store Store, other *Item, visited map[*Item]bool) (hops int32) {
	if visited[other] {
		return -1
	}

	visited[n] = true

	if n == other {
		return 0
	}

	for _, d := range n.DependsOn {
		dn := store.GetItem(d)
		if !visited[dn] {

			h := dn.isDependingOn(store, other, visited)
			if h == 0 {
				return 1
			}

			if h > 0 {
				return h + 1
			}
		}
	}

	return -1

}

// IsDependingOn checks if n depends on other and returns the number of hops
// 0 is returned if n == other; -1 is returned if n does not depend other
func (n *Item) IsDependingOn(store Store, other *Item) (hops int32) {
	return n.isDependingOn(store, other, map[*Item]bool{})
}

// AddDependency does nothing if the dependency item has the same name as the current item
func (n *Item) AddDependency(d *Item) {
	if n.Name != d.Name {
		n.DependsOn = append(n.DependsOn, d.Name)
	}
}

func (n *Item) RemoveDependency(itemName string) {
	var a []string

	for _, d := range n.DependsOn {
		if d != itemName {
			a = append(a, d)
		}
	}

	n.DependsOn = a
}

func (n *Item) AddTag(t *Tag) {
	n.Tags = append(n.Tags, t.Name)
}

func (n *Item) RemoveTag(tagName string) {
	var a []string

	for _, t := range n.Tags {
		if t != tagName {
			a = append(a, t)
		}
	}

	n.Tags = a
}

var renameLock sync.Mutex

func RenameItem(s Store, oldName, newName string) {
	renameLock.Lock()
	defer renameLock.Unlock()
	old := s.GetItem(oldName)
	s.RemoveItem(oldName, false)
	nu := s.GetItem(newName)
	*nu = *old
	nu.Name = newName
	s.EachItem(func(n *Item) {
		for i, d := range n.DependsOn {
			if d == oldName {
				n.DependsOn[i] = newName
			}
		}
	})
}

func RenameTag(s Store, oldName, newName string) {
	renameLock.Lock()
	defer renameLock.Unlock()
	old := s.GetTag(oldName)
	s.RemoveTag(oldName, false)
	nu := s.GetTag(newName)
	*nu = *old
	nu.Name = newName
	s.EachTag(func(t *Tag) {
		for i, d := range t.DependsOn {
			if d == oldName {
				t.DependsOn[i] = newName
			}
		}
	})
	s.EachItem(func(n *Item) {
		for i, tt := range n.Tags {
			if tt == oldName {
				n.Tags[i] = newName
			}
		}
	})
}

func GetItemsForTags(store Store, tags ...string) (items []*Item) {
	store.EachItem(func(n *Item) {
		var add bool
		for _, t := range n.Tags {
			for _, tt := range tags {
				if t == tt {
					add = true
				}
			}
		}

		if add {
			items = append(items, n)
		}
	})
	return
}

type wantedItem struct {
	noWanted int32
	item     *Item
}

type wantedItems []*wantedItem

func (w wantedItems) Less(i, j int) bool {
	return w[i].noWanted > w[j].noWanted
}

func (w wantedItems) Swap(i, j int) {
	w[j], w[i] = w[i], w[j]
}

func (w wantedItems) Len() int {
	return len(w)
}

/*
func printItemMap(m map[*Item]int32) {
	for n, i := range m {
		fmt.Printf("%s => %d\n", n.Name, i)
	}
}

func printWantedItems(wn wantedItems) {
	for i, wnd := range wn {
		fmt.Printf("%d. %s (%d)\n", i+1, wnd.item.Name, wnd.noWanted)
	}
}
*/

var _ = fmt.Printf

/*
func calculateMostWanted(store Store, visited map[*Item]bool, wanted int32, current *Item) int32 {
	if visited[current] {
		return wanted
	}

	for _, d := range current.DependsOn {
		dn := store.GetItem(d)
		visited[dn] = true
		wanted++
		wanted += calculateMostWanted(store, visited, 0, dn)
	}
	return wanted
}
*/

// respects that items may be wanted via other items
// so a item is wanted by all items that depend on him and all items that depend on the items that
// depend on him and so on
func GetMostWantedItems(store Store) (items []*Item) {
	// printItemMap(m)
	var wn = getMostWantedItems(store)

	// printWantedItems(wn)
	sort.Sort(wn)
	// printWantedItems(wn)

	for _, wnd := range wn {
		items = append(items, wnd.item)
	}

	return

}

func getMostWantedItems(store Store) (wn wantedItems) {
	var m = map[*Item]int32{}
	store.EachItem(func(n *Item) {
		m[n] = 0
	})

	for outer := range m {
		for inner := range m {
			if inner.IsDependingOn(store, outer) > 0 {
				m[outer]++
			}
		}
	}

	for n, no := range m {
		wn = append(wn, &wantedItem{no, n})
	}
	return
}

func getMostWantedTags(store Store) (wt wantedTags) {
	var m = map[*Tag]int32{}
	store.EachTag(func(t *Tag) {
		m[t] = 0
	})

	for outer := range m {
		for inner := range m {
			if inner.IsDependingOn(store, outer) > 0 {
				m[outer]++
			}
		}
	}

	for t, no := range m {
		wt = append(wt, &wantedTag{no, t})
	}

	sort.Sort(wt)

	return

}

type wantedTag struct {
	noWanted int32
	tag      *Tag
}

type wantedTags []*wantedTag

func (w wantedTags) Less(i, j int) bool {
	return w[i].noWanted > w[j].noWanted
}

func (w wantedTags) Swap(i, j int) {
	w[j], w[i] = w[i], w[j]
}

func (w wantedTags) Len() int {
	return len(w)
}

func addNode(m map[string]*Node, name string) {
	if _, has := m[name]; has {
		return
	}
	m[name] = &Node{Name: name}
}

func getNode(m map[string]*Node, name string) *Node {
	addNode(m, name)
	return m[name]
}

func ItemTree(store Store) *Node {
	items := getMostWantedItems(store)

	var top Node

	var nodes = map[string]*Node{}

	//maxWantedFrom := items[0].noWanted

	for _, i := range items {
		n := getNode(nodes, i.item.Name)
		n.Weight = int(i.noWanted)
		if len(i.item.DependsOn) == 0 {
			top.Children = append(top.Children, n)
		} else {
			for _, d := range i.item.DependsOn {
				dn := getNode(nodes, d)
				dn.Children = append(dn.Children, n)
			}

		}
	}
	return &top
}

func TagTree(store Store) *Node {
	tags := getMostWantedTags(store)

	var top Node

	var nodes = map[string]*Node{}

	//maxWantedFrom := tags[0].noWanted

	for _, i := range tags {
		n := getNode(nodes, i.tag.Name)
		n.Weight = int(i.noWanted)
		if len(i.tag.DependsOn) == 0 {
			top.Children = append(top.Children, n)
		} else {
			for _, d := range i.tag.DependsOn {
				dn := getNode(nodes, d)
				dn.Children = append(dn.Children, n)
			}

		}
	}
	return &top
}

func GetMostWantedTags(store Store) (tags []*Tag) {

	var wt = getMostWantedTags(store)

	for _, wtg := range wt {
		tags = append(tags, wtg.tag)
	}

	return

}

type Node struct {
	Name     string
	Weight   int
	Children []*Node
}

func walkGraph(g *gographviz.Graph, parent *Node, nodes map[string]bool) {
	for _, c := range parent.Children {
		if !nodes[c.Name] {
			var color string
			var fontcolor string
			switch w := c.Weight; w {
			case 3:
				color = "yellow"
				fontcolor = "black"
			case 4:
				color = "green"
				fontcolor = "black"
			case 5:
				color = "lightblue"
				fontcolor = "black"
			case 6:
				color = "blue"
				fontcolor = "white"
			case 7:
				color = "magenta"
				fontcolor = "white"
			default:
				if w > 8 {
					color = "red"
					fontcolor = "white"
				} else {
					color = "grey"
					fontcolor = "black"
				}

			}

			g.AddNode("G", c.Name, map[string]string{
				"shape":     "box",
				"style":     "filled",
				"fontsize":  "16",
				"color":     color,
				"fontcolor": fontcolor,
			})
			nodes[c.Name] = true
		}
		if parent.Name != "" {
			// fmt.Printf("adding Edge: %#v => %#v\n", c.Name, parent.Name)
			g.AddEdge(c.Name, parent.Name, true, map[string]string{
				"weight":    fmt.Sprintf("%d", c.Weight),
				"arrowsize": "0.6",
			})
		}

		if len(c.Children) > 0 {
			walkGraph(g, c, nodes)
		}
	}
}

func MakeGraphviz(tree *Node) string {
	g := gographviz.NewGraph()
	g.SetName("G")
	g.SetDir(true)
	g.SetStrict(false)
	g.AddAttr("G", "concentrate", "true")
	// g.AddAttr("", "ratio", "fill")
	// g.AddAttr("", "size", "16,9")
	g.AddAttr("G", "nodesep", "0.5")
	g.AddAttr("G", "ranksep", `"0.3 equally"`)
	//g.AddAttr("G", "rankdir", "RL") // or BT
	g.AddAttr("G", "rankdir", "BT") // or BT
	walkGraph(g, tree, map[string]bool{})
	return g.String()
}

/*
digraph G  {
concentrate=true
ratio=fill
size="16,9"
nodesep=0.5
ranksep="0.3 equally"
edge[arrowsize=0.6]
node [shape=box,style=filled fontsize=16 color=lightblue]
 a0 -> a1 ;
a2 -> a3;
 b0 -> b1 ;
 a1 -> b3;
  b2 -> a3;
c2 -> b2;
c2 -> a2;
b1 -> a1;
b3 -> b1;
a0 -> b2;
rankdir=RL;

b3[color=red fontcolor=white fontsize=18,label="roter Baron\n(tag1  tag 2)"]

}
*/

// node for visjs.org
type VisNode struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Value int    `json:"value,omitempty"`
	Title string `json:"title,omitempty"`
	Group string `json:"group,omitempty"`
}

// edge for visjs.org
type VisEdge struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// dataset for visjs.org
type VisDataSet struct {
	Nodes []VisNode `json:"nodes"`
	Edges []VisEdge `json:"edges"`
}

/*
func ItemTree(store Store) *Node {
	items := getMostWantedItems(store)

	var top Node

	var nodes = map[string]*Node{}

	//maxWantedFrom := items[0].noWanted

	for _, i := range items {
		n := getNode(nodes, i.item.Name)
		n.Weight = int(i.noWanted)
		if len(i.item.DependsOn) == 0 {
			top.Children = append(top.Children, n)
		} else {
			for _, d := range i.item.DependsOn {
				dn := getNode(nodes, d)
				dn.Children = append(dn.Children, n)
			}

		}
	}
	return &top
}
*/

func MakeItemsVisDataSet(store Store, tree *Node) VisDataSet {
	var vd VisDataSet

	items := getMostWantedItems(store)
	nodesNames := make(map[string]int)
	edges := [][2]string{}
	next := 1
	max := 0
	for _, item := range items {
		next++
		var vn VisNode
		vn.ID = next
		vn.Label = item.item.Name
		vn.Value = int(item.noWanted)
		vn.Title = strings.Join(item.item.Tags, ", ")
		vd.Nodes = append(vd.Nodes, vn)
		nodesNames[item.item.Name] = vn.ID
		if vn.Value > max {
			max = vn.Value
		}

		for _, d := range item.item.DependsOn {
			edges = append(edges, [2]string{item.item.Name, d})
		}
	}

	/*

			max 1200
		  groupSteps: 240

		  group 0 = 0-239
		  group 1 = 240-479
		  usw


	*/

	groupSteps := float64(max) / float64(5)

	for i, vn := range vd.Nodes {
		switch v := FloatToInt(float64(vn.Value) / groupSteps); v {
		case 0:
			vn.Group = "group0"
		case 1:
			vn.Group = "group1"
		case 2:
			vn.Group = "group2"
		case 3:
			vn.Group = "group3"
		case 4:
			vn.Group = "group4"
		case 5:
			vn.Group = "group5"
		default:
			vn.Group = "group0"
			//panic(fmt.Sprintf("should not happen, group id is %#v", v))
		}
		vd.Nodes[i] = vn
	}

	for _, e := range edges {
		var ve VisEdge
		ve.From = nodesNames[e[0]]
		ve.To = nodesNames[e[1]]
		vd.Edges = append(vd.Edges, ve)
	}

	return vd
}

func FloatToInt(x float64) int {
	return int(RoundFloat(x, 0))
}

// RoundFloat rounds the given float by the given decimals after the dot
func RoundFloat(x float64, decimals int) float64 {
	// return roundFloat(x, numDig(x)+decimals)
	frep := strconv.FormatFloat(x, 'f', decimals, 64)
	f, _ := strconv.ParseFloat(frep, 64)
	return f
}

/*
// create an array with nodes
  var nodes = new vis.DataSet([
    {id: "a", "label": "universalsprache", "value": 0, "title": "ideen, IT"},
    {id: "b", "label": "künstliche intelligenz", "value": 1, "title": "ideen, IT"},
    {id: "c", "label": "der Riese", "value": 2, "title": "zufällig"},
    {id: "d", "label": "unverbunden", "value": 0},
    {id: "e", "label": "ein bisschen", "value": 1, "title": "zufällig"}
  ]);

  // create an array with edges
  var edges = new vis.DataSet([
    {from: "a", to: "c"},
    {from: "a", to: "b"},
    {from: "b", to: "c"},
    {from: "b", to: "e"}
  ]);

*/
