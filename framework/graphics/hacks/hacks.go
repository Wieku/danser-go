package hacks

// Intel has some weird problems with DSA glBufferSubData that causes Draw calls not to be flushed immediately,
// resulting in broken image in some cases
var IsIntel bool
