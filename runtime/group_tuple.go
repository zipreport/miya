package runtime

// GroupTuple is one result of the groupby filter: a (grouper, list) pair.
// It unpacks and indexes like a two-item list and also has the attributes
// grouper and list, like Jinja2's namedtuple.
type GroupTuple []interface{}
