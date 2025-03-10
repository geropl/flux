package debug


// pass will pass any incoming tables directly next to the following transformation.
// It is best used to interrupt any planner rules that rely on a specific ordering.
builtin pass : (<-tables: [A]) => [A] where A: Record

// opaque works like `pass` in that it passes any incoming tables directly to the
// following transformation, save for its type signature does not indicate that the
// input type has any correlation with the ouput type.
builtin opaque : (<-tables: [A]) => [B] where A: Record, B: Record

// slurp will read the incoming tables and concatenate buffers with the same group key
// into a single table. This is useful for testing the performance impact of multiple
// buffers versus a single buffer.
builtin slurp : (<-tables: [A]) => [A] where A: Record

// sink will discard all data that comes into it.
builtin sink : (<-tables: [A]) => [A] where A: Record

builtin getOption : (pkg: string, name: string) => A

option vectorize = false
