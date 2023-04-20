package lru

import "container/list"

// LRU缓存策略
// 队尾是最近使用的

// 允许缓存中存储各种类型的数据，为了保证通用性，这里可以定义一个接口，只要实现了这个接口的类型都可以存储
type Value interface {
	Len() int
}

// 存储到队列中的节点
type node struct {
	key   string
	value Value
}

type HandleFunc func(string, Value)

// 缓存
type Cache struct {
	// map真正存储数据的
	cache map[string]*list.Element
	// 双端队列/链表，记录最近使用的；队尾存放的是最近使用过的节点，队头存放的是最近不使用的节点
	// queue中存储的是node
	queue *list.List
	// 最大容量
	maxCapacity int64
	// 可用容量
	availableCapacity int64
	// 已经使用的容量
	usedCapacity int64

	// 记录删除时，回调
	delete HandleFunc
	// 记录添加时，回调
	add HandleFunc
	// 记录更新时，回调
	update HandleFunc
}

func New(maxCapacity int64) *Cache {
	return &Cache{
		cache:       make(map[string]*list.Element),
		queue:       list.New(),
		maxCapacity: maxCapacity,
	}
}

// 查找
func (c *Cache) Get(key string) (value Value, ok bool) {
	// 1. 从map中查找数据; 2. 移动节点到队尾
	if element, ok := c.cache[key]; ok {
		// 移动到队尾
		c.queue.MoveToBack(element)

		node := element.Value.(*node)
		return node.value, true
	}
	return nil, false
}

// 删除最近最少使用的元素，即队头元素
func (c *Cache) RemoveOldElement() {
	// 1. 找到队头元素； 2. 从map中删除； 3. 更新所占内存； 4. 回调删除方法
	oldElement := c.queue.Front()
	if oldElement != nil {
		c.queue.Remove(oldElement)

		node := oldElement.Value.(*node)
		delete(c.cache, node.key)

		length := int64(len(node.key) + node.value.Len())
		c.usedCapacity -= length
		c.availableCapacity = c.maxCapacity - c.usedCapacity

		if c.delete != nil {
			c.delete(node.key, node.value)
		}
	}
}

// 新增、修改
func (c *Cache) Add(key string, value Value) {
	// 0. 先判断有没有；
	// 1. 添加元素到map； 2. 将节点插入到队尾； 3. 更新所占内存； 4. 回调添加方法；
	// 5. 如果内存超出最大限制，需要将最近最少使用的节点删除
	if element, ok := c.cache[key]; ok {
		c.queue.MoveToBack(element)

		node := element.Value.(*node)

		c.usedCapacity += int64(value.Len() - node.value.Len())
		c.availableCapacity = c.maxCapacity - c.usedCapacity

		node.value = value

		if c.update != nil {
			c.update(key, value)
		}

	} else {
		node := &node{key: key, value: value}
		element := c.queue.PushBack(node)

		c.cache[key] = element

		c.usedCapacity += int64(len(node.key) + node.value.Len())
		c.availableCapacity = c.maxCapacity - c.usedCapacity

		if c.add != nil {
			c.add(key, value)
		}

	}

	if c.maxCapacity > 0 && c.usedCapacity > c.maxCapacity {
		c.RemoveOldElement()
	}

}

func (c *Cache) SetDeleteHandler(handler HandleFunc) {
	c.delete = handler
}

func (c *Cache) SetAddHandler(handler HandleFunc) {
	c.add = handler
}

func (c *Cache) SetUpdateHandler(handler HandleFunc) {
	c.update = handler
}
