package transformation

// MapData is an in-memory AttrData backed by a map. Framework resources convert
// typed plan/state models into a map, run the shared mapping functions in this
// package, then read Values() back to rebuild the typed model.
type MapData struct {
	id     string
	values map[string]interface{}
}

// Ensure MapData satisfies AttrData.
var _ AttrData = (*MapData)(nil)

// NewMapData returns a MapData backed by the given values. A nil map is
// initialized to an empty one.
func NewMapData(values map[string]interface{}) *MapData {
	if values == nil {
		values = map[string]interface{}{}
	}
	return &MapData{values: values}
}

func (m *MapData) Id() string                     { return m.id }
func (m *MapData) SetId(id string)                { m.id = id }
func (m *MapData) Values() map[string]interface{} { return m.values }

func (m *MapData) Get(key string) interface{} {
	return m.values[key]
}

// GetOk mimics terraform-plugin-sdk ResourceData.GetOk: false when the key is
// absent or the value is the type's zero value. Used by ListOptionsFromData.
func (m *MapData) GetOk(key string) (interface{}, bool) {
	v, ok := m.values[key]
	if !ok {
		return v, false
	}

	switch vv := v.(type) {
	case nil:
		return v, false
	case string:
		return vv, len(vv) > 0
	case bool:
		return vv, vv
	case int:
		return vv, vv != 0
	case []interface{}:
		return vv, len(vv) > 0
	case []string:
		return vv, len(vv) > 0
	default:
		return v, true
	}
}

func (m *MapData) Set(key string, value interface{}) error {
	m.values[key] = value
	return nil
}
