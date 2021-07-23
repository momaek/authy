package service

// Searcher
type Searcher struct {
	isAfred bool
	keyword string
	*Device
}

func (s *Searcher) showAll() bool {
	return len(s.keyword) == 0
}

// NewSearcher ..
func NewSearcher(keyword string) *Searcher {
	return &Searcher{
		keyword: keyword,
		Device:  NewDevice(NewDeviceConfig{}),
	}
}

func (s *Searcher) Search() {
	s.Device.LoadTokenFromCache()
}
