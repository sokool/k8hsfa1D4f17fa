package rollee

import "sync"

type transformer func(int, int) int

type folder struct {
	init      int
	transform transformer
	result    map[ID]int

	mu sync.RWMutex
	wg sync.WaitGroup
}

func newFolder(n int, fn transformer) *folder {
	return &folder{
		init:      n,
		transform: fn,
		result:    map[ID]int{},
	}
}

func (f *folder) fold(l List) *folder {
	f.mu.Lock()
	defer f.mu.Unlock()

	n := f.init
	if v, ok := f.result[l.ID]; ok {
		n = v
	}

	for i := range l.Values {
		n = f.transform(n, l.Values[i])
	}

	f.result[l.ID] = n
	return f
}

func (f *folder) channel(c chan List) *folder {
	var wg sync.WaitGroup
	for l := range c {
		wg.Add(1)
		go func(l List) { defer wg.Done(); f.fold(l) }(l)
	}
	wg.Wait()
	return f
}

func (f *folder) slice(s []chan List) *folder {
	var wg sync.WaitGroup
	for i := range s {
		wg.Add(1)
		go func(c chan List) { defer wg.Done(); f.channel(c) }(s[i])
	}
	wg.Wait()
	return f
}
