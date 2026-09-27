package transformation

// AttrData is a view of Terraform resource/data-source attributes used by the
// shared mapping functions in this package. MapData implements it for Plugin
// Framework state.
type AttrData interface {
	Id() string
	SetId(id string)
	Get(key string) interface{}
	GetOk(key string) (interface{}, bool)
	Set(key string, value interface{}) error
}

// asInterfaceList normalizes values that implement List() and plain slices
// into []interface{}.
func asInterfaceList(v interface{}) ([]interface{}, bool) {
	switch vv := v.(type) {
	case []interface{}:
		return vv, true
	case []string:
		out := make([]interface{}, len(vv))
		for i, s := range vv {
			out[i] = s
		}
		return out, true
	case interface{ List() []interface{} }:
		return vv.List(), true
	default:
		return nil, false
	}
}
