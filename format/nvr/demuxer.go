package nvr

type DeMuxer struct {
}

//NewDeMuxer func
func NewDeMuxer() *DeMuxer {
	return &DeMuxer{}
}

//ReadIndex func
func (obj *DeMuxer) ReadIndex() (err error) {
	return nil
}

//ReadRange func
func (obj *DeMuxer) ReadRange() (err error) {
	return nil
}

//ReadGop func
func (obj *DeMuxer) ReadGop() (err error) {
	return nil
}
