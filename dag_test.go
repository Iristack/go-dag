package dag

import (
	"testing"
)

func TestNewDAG(t *testing.T) {
	d := NewDAG[string]()
	if d == nil {
		t.Fatal("NewDAG() returned nil")
	}
	if d.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes, got %d", d.NodeCount())
	}
}

func TestAddNode(t *testing.T) {
	d := NewDAG[string]()
	d.AddNode("a")
	d.AddNode("b")

	if !d.HasNode("a") {
		t.Error("Node 'a' should exist")
	}
	if d.NodeCount() != 2 {
		t.Errorf("Expected 2 nodes, got %d", d.NodeCount())
	}
}

func TestAddEdge(t *testing.T) {
	d := NewDAG[string]()
	err := d.AddEdge("a", "b")
	if err != nil {
		t.Errorf("AddEdge failed: %v", err)
	}

	if !d.HasNode("a") || !d.HasNode("b") {
		t.Error("Nodes should exist after AddEdge")
	}
	if d.EdgeCount() != 1 {
		t.Errorf("Expected 1 edge, got %d", d.EdgeCount())
	}
}

func TestAddEdgeCycle(t *testing.T) {
	d := NewDAG[string]()
	_ = d.AddEdge("a", "b")
	_ = d.AddEdge("b", "c")

	// 尝试添加会形成环的边
	err := d.AddEdge("c", "a")
	if err == nil {
		t.Error("Expected cycle detection error")
	}
}

func TestSort(t *testing.T) {
	d := NewDAG[string]()
	edges := [][2]string{
		{"a", "b"},
		{"b", "c"},
		{"c", "d"},
	}

	for _, edge := range edges {
		if err := d.AddEdge(edge[0], edge[1]); err != nil {
			t.Fatalf("AddEdge failed: %v", err)
		}
	}

	sorted, err := d.Sort()
	if err != nil {
		t.Fatalf("Sort failed: %v", err)
	}

	if len(sorted) != 4 {
		t.Errorf("Expected 4 nodes in sorted result, got %d", len(sorted))
	}

	// 验证拓扑顺序
	order := make(map[string]int)
	for i, node := range sorted {
		order[node] = i
	}

	if order["a"] > order["b"] {
		t.Error("Invalid topological order: 'a' should come before 'b'")
	}
	if order["b"] > order["c"] {
		t.Error("Invalid topological order: 'b' should come before 'c'")
	}
	if order["c"] > order["d"] {
		t.Error("Invalid topological order: 'c' should come before 'd'")
	}
}

func TestSortWithCycle(t *testing.T) {
	d := NewDAG[string]()
	_ = d.AddEdge("a", "b")
	_ = d.AddEdge("b", "c")

	// AddEdge 会阻止创建环，所以直接操作数据结构来创建环
	// 这模拟了从外部加载数据时可能存在环的情况
	d.adjacency["c"] = []string{"a"} // 直接添加边形成环
	d.reverseAdj["a"] = append(d.reverseAdj["a"], "c")

	_, err := d.Sort()
	if err == nil {
		t.Error("Expected cycle detection error in Sort")
	}
}

func TestGetDirectParents(t *testing.T) {
	d := NewDAG[string]()
	_ = d.AddEdge("a", "c")
	_ = d.AddEdge("b", "c")
	_ = d.AddEdge("d", "c")

	parents := d.GetDirectParents("c")
	if len(parents) != 3 {
		t.Errorf("Expected 3 direct parents, got %d", len(parents))
	}
}

func TestGetAllParents(t *testing.T) {
	d := NewDAG[string]()
	_ = d.AddEdge("a", "b")
	_ = d.AddEdge("b", "c")
	_ = d.AddEdge("c", "d")

	parents := d.GetAllParents("d")
	if len(parents) != 3 {
		t.Errorf("Expected 3 ancestors for 'd', got %d: %v", len(parents), parents)
	}
}

func TestGetChildren(t *testing.T) {
	d := NewDAG[string]()
	_ = d.AddEdge("a", "b")
	_ = d.AddEdge("a", "c")

	children := d.GetChildren("a")
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

func TestGetAllChildren(t *testing.T) {
	d := NewDAG[string]()
	_ = d.AddEdge("a", "b")
	_ = d.AddEdge("b", "c")
	_ = d.AddEdge("c", "d")

	children := d.GetAllChildren("a")
	if len(children) != 3 {
		t.Errorf("Expected 3 descendants for 'a', got %d: %v", len(children), children)
	}
}

func TestHasPath(t *testing.T) {
	d := NewDAG[string]()
	_ = d.AddEdge("a", "b")
	_ = d.AddEdge("b", "c")

	if !d.HasPath("a", "c") {
		t.Error("Expected path from 'a' to 'c'")
	}
	if d.HasPath("c", "a") {
		t.Error("Should not have path from 'c' to 'a'")
	}
}

func TestMultipleParents(t *testing.T) {
	d := NewDAG[string]()
	edges := [][2]string{
		{"aaa", "b"},
		{"aa", "b"},
		{"a", "b"},
		{"b", "c"},
		{"c", "d"},
		{"d", "e"},
	}

	for _, edge := range edges {
		if err := d.AddEdge(edge[0], edge[1]); err != nil {
			t.Fatalf("AddEdge failed: %v", err)
		}
	}

	// 测试直接父节点
	directParents := d.GetDirectParents("b")
	if len(directParents) != 3 {
		t.Errorf("Expected 3 direct parents for 'b', got %d: %v", len(directParents), directParents)
	}

	// 测试所有祖先
	allParents := d.GetAllParents("c")
	if len(allParents) != 4 {
		t.Errorf("Expected 4 ancestors for 'c', got %d: %v", len(allParents), allParents)
	}

	// 测试拓扑排序
	sorted, err := d.Sort()
	if err != nil {
		t.Fatalf("Sort failed: %v", err)
	}
	if len(sorted) != 7 {
		t.Errorf("Expected 7 nodes in sorted result, got %d", len(sorted))
	}
}

// TestIntDAG 测试整数类型的 DAG
func TestIntDAG(t *testing.T) {
	d := NewDAG[int]()

	// 创建图：1 -> 2 -> 3 -> 4
	edges := [][2]int{
		{1, 2},
		{2, 3},
		{3, 4},
	}

	for _, edge := range edges {
		if err := d.AddEdge(edge[0], edge[1]); err != nil {
			t.Fatalf("AddEdge failed: %v", err)
		}
	}

	sorted, err := d.Sort()
	if err != nil {
		t.Fatalf("Sort failed: %v", err)
	}

	if len(sorted) != 4 {
		t.Errorf("Expected 4 nodes in sorted result, got %d", len(sorted))
	}

	// 验证顺序
	for i := 0; i < len(sorted)-1; i++ {
		if sorted[i]+1 != sorted[i+1] {
			t.Errorf("Expected consecutive integers, got %d -> %d", sorted[i], sorted[i+1])
		}
	}
}

// TestSortWithOrder 测试带排序函数的拓扑排序
func TestSortWithOrder(t *testing.T) {
	d := NewDAG[int]()

	// 创建图：3 -> 1, 3 -> 2
	_ = d.AddEdge(3, 1)
	_ = d.AddEdge(3, 2)

	// 使用自定义排序（升序）
	sorted, err := d.SortWithOrder(func(a, b int) bool {
		return a < b
	})
	if err != nil {
		t.Fatalf("SortWithOrder failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("Expected 3 nodes in sorted result, got %d", len(sorted))
	}

	// 3 应该在最前面
	if sorted[0] != 3 {
		t.Errorf("Expected first node to be 3, got %d", sorted[0])
	}
}

// TestStructDAG 测试结构体类型的 DAG
func TestStructDAG(t *testing.T) {
	type Node struct {
		ID   int
		Name string
	}

	d := NewDAG[Node]()

	nodes := []Node{
		{ID: 1, Name: "A"},
		{ID: 2, Name: "B"},
		{ID: 3, Name: "C"},
	}

	_ = d.AddEdge(nodes[0], nodes[1])
	_ = d.AddEdge(nodes[1], nodes[2])

	sorted, err := d.Sort()
	if err != nil {
		t.Fatalf("Sort failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("Expected 3 nodes in sorted result, got %d", len(sorted))
	}
}
