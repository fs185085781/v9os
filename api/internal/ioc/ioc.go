package ioc

import "sync"

// IocContainer 依赖注入容器
type IocContainer struct {
	container sync.Map
}

var iocContainer = &IocContainer{}
var mu sync.Mutex

func Ioc() *IocContainer {
	return iocContainer
}
func (i *IocContainer) Register(name string, obj interface{}) {
	i.container.Store(name, obj)
}
func (i *IocContainer) Get(name string) interface{} {
	obj, _ := i.container.Load(name)
	return obj
}
func (i *IocContainer) Unregister(name string) {
	i.container.Delete(name)
}
func (i *IocContainer) GetOrRegister(name string, obj interface{}) interface{} {
	a, _ := i.container.LoadOrStore(name, obj)
	return a
}
func (i *IocContainer) RegisterList(name string, obj interface{}) {
	mu.Lock()
	defer mu.Unlock()
	a, _ := i.container.LoadOrStore(name, []interface{}{})
	if list, ok := a.([]interface{}); ok {
		list = append(list, obj)
		i.container.Store(name, list)
	}
}
