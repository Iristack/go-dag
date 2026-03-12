package dag

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
)

// DAG 表示有向无环图，支持泛型节点
// T 必须是 comparable 类型，可以用作 map 的 key
type DAG[T comparable] struct {
	nodes      map[T]bool
	adjacency  map[T][]T // 正向邻接表：node -> 子节点列表
	reverseAdj map[T][]T // 反向邻接表：node -> 父节点列表
}

// NewDAG 创建一个新的 DAG
func NewDAG[T comparable]() *DAG[T] {
	return &DAG[T]{
		nodes:      make(map[T]bool),
		adjacency:  make(map[T][]T),
		reverseAdj: make(map[T][]T),
	}
}

// AddNode 添加一个节点
func (d *DAG[T]) AddNode(node T) {
	if _, exists := d.nodes[node]; !exists {
		d.nodes[node] = true
		d.adjacency[node] = []T{}
		d.reverseAdj[node] = []T{}
	}
}

// AddEdge 添加一条边 from -> to
// 如果会形成环，返回错误
func (d *DAG[T]) AddEdge(from, to T) error {
	d.AddNode(from)
	d.AddNode(to)

	// 检查边是否已存在
	for _, neighbor := range d.adjacency[from] {
		if neighbor == to {
			return nil
		}
	}

	// 检查是否会形成环
	if d.HasPath(to, from) {
		return fmt.Errorf("cycle detected: cannot add edge %v -> %v", from, to)
	}

	d.adjacency[from] = append(d.adjacency[from], to)
	d.reverseAdj[to] = append(d.reverseAdj[to], from)
	return nil
}

// HasPath 检查从 from 到 to 是否存在路径
func (d *DAG[T]) HasPath(from, to T) bool {
	return d.hasPath(from, to, make(map[T]bool))
}

func (d *DAG[T]) hasPath(from, to T, visited map[T]bool) bool {
	if from == to {
		return true
	}
	if visited[from] {
		return false
	}
	visited[from] = true

	for _, neighbor := range d.adjacency[from] {
		if d.hasPath(neighbor, to, visited) {
			return true
		}
	}
	return false
}

// sort 排序
func (d *DAG[T]) sort(queue []T, inDegree map[T]int) ([]T, error) {
	sorted := make([]T, 0, len(d.nodes))
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		for _, neighbor := range d.adjacency[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(sorted) != len(d.nodes) {
		return nil, fmt.Errorf("cycle detected in graph")
	}
	return sorted, nil
}

// Sort 拓扑排序，返回节点的线性序列
// 如果存在环，返回错误
func (d *DAG[T]) Sort() ([]T, error) {
	inDegree := make(map[T]int, len(d.nodes))

	// 初始化所有节点的入度
	for node := range d.nodes {
		inDegree[node] = 0
	}

	// 计算每个节点的入度
	for _, neighbors := range d.adjacency {
		for _, neighbor := range neighbors {
			inDegree[neighbor]++
		}
	}

	// 收集入度为 0 的节点
	queue := make([]T, 0, len(d.nodes))
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	return d.sort(queue, inDegree)
}

// SortWithOrder 拓扑排序，返回按指定顺序排序的节点序列
func (d *DAG[T]) SortWithOrder(less func(a, b T) bool) ([]T, error) {
	inDegree := make(map[T]int, len(d.nodes))

	// 初始化所有节点的入度
	for node := range d.nodes {
		inDegree[node] = 0
	}

	// 计算每个节点的入度
	for _, neighbors := range d.adjacency {
		for _, neighbor := range neighbors {
			inDegree[neighbor]++
		}
	}

	// 收集入度为 0 的节点
	queue := make([]T, 0, len(d.nodes))
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	// 使用自定义比较函数排序
	if less != nil {
		sort.Slice(queue, func(i, j int) bool {
			return less(queue[i], queue[j])
		})
	}

	return d.sort(queue, inDegree)
}

// GetChildren 获取某个节点的直接子节点
func (d *DAG[T]) GetChildren(node T) []T {
	return d.adjacency[node]
}

// GetDirectParents 获取某个节点的直接父节点
func (d *DAG[T]) GetDirectParents(node T) []T {
	return d.reverseAdj[node]
}

// GetAllParents 获取某个节点的所有上级节点（祖先节点）
func (d *DAG[T]) GetAllParents(node T) []T {
	parents := make([]T, 0)
	visited := make(map[T]bool)
	d.collectAllParents(node, &parents, visited)
	return parents
}

func (d *DAG[T]) collectAllParents(node T, parents *[]T, visited map[T]bool) {
	if visited[node] {
		return
	}
	visited[node] = true

	for _, parent := range d.reverseAdj[node] {
		if !visited[parent] {
			*parents = append(*parents, parent)
			d.collectAllParents(parent, parents, visited)
		}
	}
}

// GetAllChildren 获取某个节点的所有下级节点（子孙节点）
func (d *DAG[T]) GetAllChildren(node T) []T {
	children := make([]T, 0)
	visited := make(map[T]bool)
	d.collectAllChildren(node, &children, visited)
	return children
}

func (d *DAG[T]) collectAllChildren(node T, children *[]T, visited map[T]bool) {
	if visited[node] {
		return
	}
	visited[node] = true

	for _, child := range d.adjacency[node] {
		if !visited[child] {
			*children = append(*children, child)
			d.collectAllChildren(child, children, visited)
		}
	}
}

// GetNodes 获取所有节点
func (d *DAG[T]) GetNodes() []T {
	nodes := make([]T, 0, len(d.nodes))
	for node := range d.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// HasNode 检查节点是否存在
func (d *DAG[T]) HasNode(node T) bool {
	return d.nodes[node]
}

// NodeCount 返回节点数量
func (d *DAG[T]) NodeCount() int {
	return len(d.nodes)
}

// EdgeCount 返回边数量
func (d *DAG[T]) EdgeCount() int {
	count := 0
	for _, neighbors := range d.adjacency {
		count += len(neighbors)
	}
	return count
}

// dagDto 用于序列化的数据传输对象
type dagDto[T comparable] struct {
	Nodes      []T       `json:"nodes"`
	Edges      []edge[T] `json:"edges"`
	Adjacency  map[T][]T `json:"adjacency,omitempty"`
	ReverseAdj map[T][]T `json:"reverse_adj,omitempty"`
}

// edge 表示一条边
type edge[T comparable] struct {
	From T `json:"from"`
	To   T `json:"to"`
}

// Serialize 序列化图为 JSON 字节
func (d *DAG[T]) Serialize() ([]byte, error) {
	// 收集所有节点
	nodes := make([]T, 0, len(d.nodes))
	for node := range d.nodes {
		nodes = append(nodes, node)
	}

	// 收集所有边
	edges := make([]edge[T], 0)
	for from, neighbors := range d.adjacency {
		for _, to := range neighbors {
			edges = append(edges, edge[T]{From: from, To: to})
		}
	}

	dto := dagDto[T]{
		Nodes: nodes,
		Edges: edges,
	}

	return json.Marshal(dto)
}

// SerializeWithAdjacency 序列化图为 JSON 字节(仅邻接表)
func (d *DAG[T]) SerializeWithAdjacency() ([]byte, error) {
	dto := dagDto[T]{
		Adjacency:  d.adjacency,
		ReverseAdj: d.reverseAdj,
	}

	return json.Marshal(dto)
}

// Deserialize 从 JSON 字节反序列化图
func (d *DAG[T]) Deserialize(data []byte) error {
	var dto dagDto[T]
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}

	// 重置图
	d.nodes = make(map[T]bool)
	d.adjacency = make(map[T][]T)
	d.reverseAdj = make(map[T][]T)

	// 添加所有节点
	for _, node := range dto.Nodes {
		d.AddNode(node)
	}

	// 添加所有边
	for _, edge := range dto.Edges {
		if err := d.AddEdge(edge.From, edge.To); err != nil {
			return fmt.Errorf("failed to add edge %v -> %v: %w", edge.From, edge.To, err)
		}
	}

	return nil
}

// DeserializeWithAdjacency 从 JSON 字节反序列化图
func (d *DAG[T]) DeserializeWithAdjacency(data []byte) error {
	var dto dagDto[T]
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}

	// 重置图
	d.nodes = make(map[T]bool)
	d.adjacency = make(map[T][]T)
	d.reverseAdj = make(map[T][]T)

	// 直接从邻接表恢复
	d.adjacency = dto.Adjacency
	d.reverseAdj = dto.ReverseAdj

	// 从邻接表重建节点集合
	for node := range d.adjacency {
		d.nodes[node] = true
	}
	for _, neighbors := range d.adjacency {
		for _, neighbor := range neighbors {
			d.nodes[neighbor] = true
		}
	}

	return nil
}

// SerializeToBase64 序列化图为 Base64 字符串
func (d *DAG[T]) SerializeToBase64() (string, error) {
	data, err := d.Serialize()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// DeserializeFromBase64 从 Base64 字符串反序列化图
func (d *DAG[T]) DeserializeFromBase64(data string) error {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	return d.Deserialize(decoded)
}

// GetLayers 获取 DAG 的分层结构
// 返回按层级分组的节点列表，第一层是入度为 0 的节点（根节点）
// 每一层包含所有直接子节点
func (d *DAG[T]) GetLayers() [][]T {
	if len(d.nodes) == 0 {
		return [][]T{}
	}

	// 计算每个节点的入度
	inDegree := make(map[T]int, len(d.nodes))
	for node := range d.nodes {
		inDegree[node] = 0
	}
	for _, neighbors := range d.adjacency {
		for _, neighbor := range neighbors {
			inDegree[neighbor]++
		}
	}

	// 找到所有入度为 0 的节点（第一层）
	currentLayer := make([]T, 0)
	for node, degree := range inDegree {
		if degree == 0 {
			currentLayer = append(currentLayer, node)
		}
	}

	if len(currentLayer) == 0 {
		// 图中存在环
		return [][]T{}
	}

	layers := [][]T{currentLayer}
	visited := make(map[T]bool)

	// 逐层遍历
	for len(currentLayer) > 0 {
		nextLayerMap := make(map[T]bool)

		for _, node := range currentLayer {
			visited[node] = true
			for _, child := range d.adjacency[node] {
				if !visited[child] {
					// 检查是否所有父节点都已访问
					allParentsVisited := true
					for _, parent := range d.reverseAdj[child] {
						if !visited[parent] {
							allParentsVisited = false
							break
						}
					}
					if allParentsVisited {
						nextLayerMap[child] = true
					}
				}
			}
		}

		if len(nextLayerMap) == 0 {
			break
		}

		nextLayer := make([]T, 0, len(nextLayerMap))
		for node := range nextLayerMap {
			nextLayer = append(nextLayer, node)
		}
		layers = append(layers, nextLayer)
		currentLayer = nextLayer
	}

	return layers
}

// GetLayerJSON 获取 DAG 的分层结构并返回 JSON
func (d *DAG[T]) GetLayerJSON() ([]byte, error) {
	layers := d.GetLayers()
	return json.Marshal(layers)
}
