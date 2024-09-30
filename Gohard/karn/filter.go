package karn

const (	
	FilterTypeEQ = "eq"
)

func eq(a, b any) bool {
	return a == b
}

type comparison func (a, b any) bool

type compFilter struct {
		kvs Map
		comp comparison

}
	
func (f compFilter) apply(record Map) bool {
		for k, v := range f.kvs {
			value, ok := record[k]
			if !ok {
				return false
			}
			if k == "id" {
				return f.comp(value, uint64(v.(int)))
			}
			return f.comp(value, v)
		}
		return true

}

type Filter struct {
		karn 	*Karn
		coll	string
		compFilter []compFilter
		slct 	[]string
		limit 	int
}

func NewFilter(db *Karn, coll string) *Filter {

	return &Filter{

		karn: db,

		coll: coll,

		compFilter: make([]compFilter, 0),

	}

}


func (f *Filter) Eq(value Map) *Filter {
		filt := compFilter{
				comp: eq,
				kvs: value,
		}
		f.compFilter = append(f.compFilter, filt)
		return f
}

//insert the values
func (f *Filter) Insert(values Map) (uint64, error) {
		tx, err := f.karn.db.Begin(true)
		if err != nil {
			return 0, err
		}
		defer tx.Rollback()

		collBucket, err := tx.CreateBucketIfNotExists([]byte(f.coll))
		if err != nil {
		return 0, err
		}
		id, err := collBucket.NextSequence()
		if err != nil {
			return 0, err
		}
		b, err := f.karn.Encoder.Encode(values)
		if err != nil {
		return 0, err
		}
		if err := collBucket.Put(uint64Bytes(id), b); err != nil {
			return 0, err
		}
		return id, tx.Commit()

	}

func (f *Filter) Update(values Map) ([]Map, error) {
		tx, err := f.karn.db.Begin(true)
}
		f.slct = fields
		return f
}