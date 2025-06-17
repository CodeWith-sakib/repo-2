package dashboard

type ViewModelFilter struct {
	StateFilter string
}

func NewViewModelFilter(state string) *ViewModelFilter {
	return &ViewModelFilter{StateFilter: state}
}

func (f *ViewModelFilter) Match(state string) bool {
	if f.StateFilter == "" {
		return true
	}
	return f.StateFilter == state
}
