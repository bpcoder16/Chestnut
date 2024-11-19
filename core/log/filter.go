package log

// FilterOption is filter option.
type FilterOption func(*Filter)

const fuzzyStr = "******"

// Filter is a zaplogger filter.
type Filter struct {
	logger Logger
	level  Level
	key    map[interface{}]struct{}
	value  map[interface{}]struct{}
	filter func(level Level, keyValues ...interface{}) bool
}

func (f *Filter) Log(level Level, keyValues ...interface{}) error {
	if level < f.level {
		return nil
	}
	// prefixKeyValues contains the slice of arguments defined as prefixes during the logit initialization
	var prefixKeyValues []interface{}
	if l, ok := f.logger.(*zLogger); ok && len(l.prefix) > 0 {
		prefixKeyValues = make([]interface{}, 0, len(l.prefix))
		prefixKeyValues = append(prefixKeyValues, l.prefix...)
	}

	if f.filter != nil && (f.filter(level, prefixKeyValues...) || f.filter(level, keyValues...)) {
		return nil
	}

	if len(f.key) > 0 || len(f.value) > 0 {
		for i := 0; i < len(keyValues); i += 2 {
			v := i + 1
			if v >= len(keyValues) {
				break
			}
			if _, ok := f.key[keyValues[i]]; ok {
				keyValues[v] = fuzzyStr
			}
			if _, ok := f.value[keyValues[v]]; ok {
				keyValues[v] = fuzzyStr
			}
		}
	}
	return f.logger.Log(level, keyValues...)
}

// NewFilter new a zaplogger filter.
func NewFilter(logger Logger, opts ...FilterOption) *Filter {
	options := Filter{
		logger: logger,
		level:  LevelDebug,
		key:    make(map[interface{}]struct{}),
		value:  make(map[interface{}]struct{}),
	}
	for _, o := range opts {
		o(&options)
	}
	return &options
}

// FilterLevel with filter level.
func FilterLevel(level Level) FilterOption {
	return func(opts *Filter) {
		opts.level = level
	}
}

// FilterKey with filter key.
func FilterKey(keys ...string) FilterOption {
	return func(o *Filter) {
		for _, v := range keys {
			o.key[v] = struct{}{}
		}
	}
}

// FilterValue with filter value.
func FilterValue(values ...string) FilterOption {
	return func(o *Filter) {
		for _, v := range values {
			o.value[v] = struct{}{}
		}
	}
}

// FilterFunc with filter func.
func FilterFunc(f func(level Level, keyValues ...interface{}) bool) FilterOption {
	return func(o *Filter) {
		o.filter = f
	}
}
