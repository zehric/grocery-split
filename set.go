package main

type StringSet struct {
	Vals map[string]struct{}
}

func (set *StringSet) Contains(s string) (ok bool) {
	_, ok = set.Vals[s]
	return
}

func (set *StringSet) Add(s string) (ok bool) {
	_, ok = set.Vals[s]
	if ok {
		return
	}
	set.Vals[s] = struct{}{}
	return
}

func (set *StringSet) Remove(s string) bool {
	_, ok := set.Vals[s]
	if ok {
		delete(set.Vals, s)
	}
	return !ok
}

func MakeSetFromSlice(strings []string) (set StringSet) {
	set = StringSet{Vals: make(map[string]struct{})}
	for _, s := range strings {
		set.Add(s)
	}
	return
}

func (set *StringSet) Length() int {
	return len(set.Vals)
}
