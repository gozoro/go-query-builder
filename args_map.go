package query_builder

import "strings"

type argsMap struct {
	opts *options
	args map[string]any

	limit          uint
	offset         uint
	orderKey       string
	defaultOrderBy string
	orderByRules   map[string]string
}

// A Option sets options such as limit Param, offset Param etc.
type Option func(*options)

type options struct {
	limitParam   string
	limitDefault uint
	limitMax     uint
	offsetParam  string
	offsetFunc   func(uint) uint
	orderParams  []string
	orderKeyFunc func([]string) string
}

var defaultOptions = options{
	limitParam:   "limit",
	limitDefault: 10,
	limitMax:     1000,
	offsetParam:  "offset",
	offsetFunc: func(offset uint) uint {
		return offset
	},
	orderParams: []string{"sort"},
	orderKeyFunc: func(params []string) string {
		return strings.Join(params, " ")
	},
}

// NewArgs parses a map of query parameters, extracts pagination and sorting controls,
// and returns an argsMap for safe SQL generation. It applies configured limits,
// transforms offsets, builds a composite order key, and removes processed params
// from the remaining args map. Accepts optional Option functions to customize behavior.
func NewArgs(args map[string]any, opt ...Option) *argsMap {

	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}

	m := &argsMap{}

	if limit, ok := args[opts.limitParam]; ok {
		m.limit = min(limit.(uint), opts.limitMax)
		delete(args, opts.limitParam)
	} else {
		m.limit = opts.limitDefault
	}

	if offset, ok := args[opts.offsetParam]; ok {
		m.offset = opts.offsetFunc(offset.(uint))
		delete(args, opts.offsetParam)
	}

	orderValues := make([]string, 0, len(opts.orderParams))
	for _, orderParam := range opts.orderParams {

		if orderVal, ok := args[orderParam]; ok {
			orderValues = append(orderValues, orderVal.(string))
			delete(args, orderParam)
		}
	}

	m.orderKey = opts.orderKeyFunc(orderValues)
	m.args = args

	return m
}

// WithLimitParam is an Option to sets the parameter name to limit
func WithLimitParam(limitParam string) Option {
	return func(o *options) {
		o.limitParam = limitParam
	}
}

// WithLimitDefault is an Option to sets the default value for limit
func WithLimitDefault(limitDefault uint) Option {
	return func(o *options) {
		o.limitDefault = limitDefault
	}
}

// WithLimitMax is an Option to sets the max value for limit
func WithLimitMax(limitMax uint) Option {
	return func(o *options) {
		o.limitMax = limitMax
	}
}

// WithOffsetParam is an Option to sets the parameter name to offset
func WithOffsetParam(offsetParam string) Option {
	return func(o *options) {
		o.offsetParam = offsetParam
	}
}

// WithOffsetFunc is an Option to sets the offset calculation function
func WithOffsetFunc(offsetFunc func(uint) uint) Option {
	return func(o *options) {
		o.offsetFunc = offsetFunc
	}
}

// WithOrderByParams is an Option to sets the names of the parameters for getting the sorting values
func WithOrderByParams(orderParams []string) Option {
	return func(o *options) {
		o.orderParams = orderParams
	}
}

// WithOrderByKeyFunc is an Option to sets the function for obtaining a string key for combining multiple sorting parameters
func WithOrderByKeyFunc(orderFunc func([]string) string) Option {
	return func(o *options) {
		o.orderKeyFunc = orderFunc
	}
}

// SetDefaultOrderBy is a method to sets default value of ORDER BY
func (m *argsMap) SetDefaultOrderBy(defaultOrderBy string) *argsMap {
	m.defaultOrderBy = defaultOrderBy
	return m
}

// AddOrderRule registers a mapping from a user-facing orderKey to a safe SQL ORDER BY clause.
//   - orderKey: the string formed by concatenating sort parameter values (e.g., "price desc").
//   - orderBy: the validated SQL expression to use (e.g., "price DESC, id ASC").
//
// Returns the argsMap for method chaining.
func (m *argsMap) AddOrderRule(orderKey, orderBy string) *argsMap {

	m.orderByRules[orderKey] = orderBy
	return m
}

// Value is a method to returns value of parameter input args
func (m *argsMap) Value(param string) any {

	if val, ok := m.args[param]; ok {
		return val
	}

	return nil
}

// Limit is a method to returns the value of the LIMIT clause in the SQL query.
func (m *argsMap) Limit() uint {
	return m.limit
}

// Offset is a method to returns the value of the OFFSET clause in the SQL query.
func (m *argsMap) Offset() uint {
	return m.offset
}

// OrderBy is a method to returns the value of the ORDER BY clause in the SQL query.
func (m *argsMap) OrderBy() string {

	if orderBy, ok := m.orderByRules[m.orderKey]; ok {
		return orderBy
	}

	return m.defaultOrderBy
}
