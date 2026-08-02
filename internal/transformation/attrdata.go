package transformation

// AttrData is an SDK-agnostic view of Terraform resource/data-source attributes.
// It decouples domain mapping from terraform-plugin-sdk's *schema.ResourceData
// so the same mapping logic can target Plugin Framework state via MapData.
//
// *schema.ResourceData (SDKv2) and *MapData (Framework bridge) both satisfy this
// interface.
type AttrData interface {
	Id() string
	SetId(id string)
	Get(key string) interface{}
	GetOk(key string) (interface{}, bool)
	Set(key string, value interface{}) error
}

// asInterfaceList normalizes SDKv2 *schema.Set values (via List()) and plain
// slices into []interface{} without importing the SDK.
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
